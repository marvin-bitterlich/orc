package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/agent"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// MailCmd returns the mail command
func MailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mail",
		Short: "Inter-agent mail system",
		Long: `Send and receive messages between ORC and IMP agents.

Messages are async and persistent in the ORC database.
Agent identity is auto-detected from context (ORC repo or grove).`,
	}

	cmd.AddCommand(mailSendCmd())
	cmd.AddCommand(mailInboxCmd())
	cmd.AddCommand(mailReadCmd())
	cmd.AddCommand(mailConversationCmd())

	return cmd
}

func mailSendCmd() *cobra.Command {
	var to, subject string

	cmd := &cobra.Command{
		Use:   "send <body>",
		Short: "Send a message to another agent",
		Long: `Send a message to ORC or an IMP agent.

Examples:
  orc mail send "Please review PR #42" --to IMP-GROVE-001 --subject "Code Review"
  orc mail send "Task complete" --to ORC --subject "Update"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			body := args[0]

			// Get current agent identity
			identity, err := agent.GetCurrentAgentID()
			if err != nil {
				return fmt.Errorf("failed to detect agent identity: %w", err)
			}

			// Validate recipient
			if to == "" {
				return fmt.Errorf("--to is required")
			}

			recipientIdentity, err := agent.ParseAgentID(to)
			if err != nil {
				return fmt.Errorf("invalid recipient: %w", err)
			}

			// Use subject or generate default
			if subject == "" {
				subject = "(no subject)"
			}

			// Determine mission ID for message
			// Messages must be scoped to a mission for database storage
			missionID := identity.MissionID

			if identity.Type == agent.AgentTypeORC {
				// ORC sending: use recipient's mission ID (must be IMP)
				if recipientIdentity.Type == agent.AgentTypeIMP {
					// Need to look up grove to get mission ID
					if recipientIdentity.MissionID == "" {
						// Try to extract from grove
						grove, err := wire.GroveService().GetGrove(ctx, recipientIdentity.ID)
						if err != nil {
							return fmt.Errorf("failed to resolve IMP mission: %w", err)
						}
						missionID = grove.MissionID
					} else {
						missionID = recipientIdentity.MissionID
					}
				} else {
					return fmt.Errorf("ORC can only send to IMP agents")
				}
			} else if recipientIdentity.Type == agent.AgentTypeORC {
				// Sending TO ORC: use sender's mission ID (IMPs reporting to ORC)
				missionID = identity.MissionID
			}
			// Otherwise: IMP to IMP, use sender's mission ID (already set)

			// Create message
			resp, err := wire.MessageService().CreateMessage(ctx, primary.CreateMessageRequest{
				Sender:    identity.FullID,
				Recipient: recipientIdentity.FullID,
				Subject:   subject,
				Body:      body,
				MissionID: missionID,
			})
			if err != nil {
				return fmt.Errorf("failed to create message: %w", err)
			}

			fmt.Printf("✓ Message sent: %s\n", resp.MessageID)
			fmt.Printf("  From: %s\n", identity.FullID)
			fmt.Printf("  To: %s\n", recipientIdentity.FullID)
			fmt.Printf("  Subject: %s\n", subject)

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Recipient agent ID (e.g., IMP-GROVE-001)")
	cmd.Flags().StringVar(&subject, "subject", "", "Message subject")
	cmd.MarkFlagRequired("to")

	return cmd
}

func mailInboxCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "View your inbox",
		Long: `List messages in your inbox.

By default, shows only unread messages.
Use --all to show all messages.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get current agent identity
			identity, err := agent.GetCurrentAgentID()
			if err != nil {
				return fmt.Errorf("failed to detect agent identity: %w", err)
			}

			// Get messages
			messages, err := wire.MessageService().ListMessages(ctx, identity.FullID, !all)
			if err != nil {
				return fmt.Errorf("failed to list messages: %w", err)
			}

			if len(messages) == 0 {
				if all {
					fmt.Println("No messages")
				} else {
					fmt.Println("No unread messages")
				}
				return nil
			}

			// Display header
			fmt.Printf("Inbox for %s\n\n", identity.FullID)

			// Display messages
			for _, msg := range messages {
				status := "✉"
				if msg.Read {
					status = "✓"
				}
				fmt.Printf("%s %s [%s]\n", status, msg.ID, msg.Timestamp)
				fmt.Printf("  From: %s\n", msg.Sender)
				fmt.Printf("  Subject: %s\n", msg.Subject)
				fmt.Printf("  Body: %s\n", truncate(msg.Body, 60))
				fmt.Println()
			}

			// Display summary
			unreadCount, _ := wire.MessageService().GetUnreadCount(ctx, identity.FullID)
			fmt.Printf("Total: %d messages (%d unread)\n", len(messages), unreadCount)

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Show all messages (not just unread)")

	return cmd
}

func mailReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <message-id>",
		Short: "Read a specific message",
		Long: `Display a message and mark it as read.

Example:
  orc mail read MSG-MISSION-001-005`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			messageID := args[0]

			// Get message
			message, err := wire.MessageService().GetMessage(ctx, messageID)
			if err != nil {
				return fmt.Errorf("failed to get message: %w", err)
			}

			// Display message
			fmt.Printf("Message: %s\n", message.ID)
			fmt.Printf("From: %s\n", message.Sender)
			fmt.Printf("To: %s\n", message.Recipient)
			fmt.Printf("Subject: %s\n", message.Subject)
			fmt.Printf("Date: %s\n", message.Timestamp)
			fmt.Printf("\n%s\n", message.Body)

			// Mark as read
			if !message.Read {
				if err := wire.MessageService().MarkRead(ctx, messageID); err != nil {
					return fmt.Errorf("failed to mark as read: %w", err)
				}
				fmt.Println("\n✓ Marked as read")
			}

			return nil
		},
	}

	return cmd
}

func mailConversationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conversation <agent-id>",
		Short: "View conversation thread with another agent",
		Long: `Display all messages between you and another agent.

Example:
  orc mail conversation IMP-GROVE-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			otherAgentID := args[0]

			// Validate other agent ID
			otherIdentity, err := agent.ParseAgentID(otherAgentID)
			if err != nil {
				return fmt.Errorf("invalid agent ID: %w", err)
			}

			// Get current agent identity
			identity, err := agent.GetCurrentAgentID()
			if err != nil {
				return fmt.Errorf("failed to detect agent identity: %w", err)
			}

			// Get conversation
			messages, err := wire.MessageService().GetConversation(ctx, identity.FullID, otherIdentity.FullID)
			if err != nil {
				return fmt.Errorf("failed to get conversation: %w", err)
			}

			if len(messages) == 0 {
				fmt.Printf("No conversation with %s\n", otherIdentity.FullID)
				return nil
			}

			// Display header
			fmt.Printf("Conversation: %s ↔ %s\n\n", identity.FullID, otherIdentity.FullID)

			// Display messages
			for _, msg := range messages {
				direction := "←"
				if msg.Sender == identity.FullID {
					direction = "→"
				}

				fmt.Printf("%s [%s] %s\n", direction, msg.Timestamp, msg.Subject)
				fmt.Printf("  %s\n", msg.Body)
				fmt.Println()
			}

			fmt.Printf("Total: %d messages\n", len(messages))

			return nil
		},
	}

	return cmd
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	// Replace newlines with spaces for display
	s = strings.ReplaceAll(s, "\n", " ")

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
