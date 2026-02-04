package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/agent"
	"github.com/example/orc/internal/config"
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
Agent identity is auto-detected from context (ORC repo or workbench).`,
	}

	cmd.AddCommand(mailSendCmd())
	cmd.AddCommand(mailInboxCmd())
	cmd.AddCommand(mailReadCmd())
	cmd.AddCommand(mailConversationCmd())

	return cmd
}

func mailSendCmd() *cobra.Command {
	var to, subject string
	var nudge, toGoblin bool

	cmd := &cobra.Command{
		Use:   "send <body>",
		Short: "Send a message to another agent",
		Long: `Send a message to ORC or an IMP agent.

Examples:
  orc mail send "Please review PR #42" --to IMP-WB-001 --subject "Code Review"
  orc mail send "Task complete" --to ORC --subject "Update"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			body := args[0]

			// Get current agent identity
			identity, err := agent.GetCurrentAgentID()
			if err != nil {
				return fmt.Errorf("failed to detect agent identity: %w", err)
			}

			// Resolve recipient
			var recipientIdentity *agent.AgentIdentity

			if toGoblin {
				// Resolve goblin from workbench context
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}

				cfg, err := config.LoadConfig(cwd)
				if err != nil || cfg.PlaceID == "" {
					return fmt.Errorf("--to-goblin requires workbench context (no .orc/config.json found)")
				}

				if !config.IsWorkbench(cfg.PlaceID) {
					return fmt.Errorf("--to-goblin requires workbench context (found %s)", cfg.PlaceID)
				}

				// Get workbench -> workshop -> gatehouse
				workbench, err := wire.WorkbenchService().GetWorkbench(ctx, cfg.PlaceID)
				if err != nil {
					return fmt.Errorf("failed to get workbench: %w", err)
				}

				gatehouse, err := wire.GatehouseService().GetGatehouseByWorkshop(ctx, workbench.WorkshopID)
				if err != nil {
					return fmt.Errorf("failed to resolve goblin: %w", err)
				}

				recipientIdentity = &agent.AgentIdentity{
					Type:   agent.AgentTypeGoblin,
					ID:     gatehouse.ID,
					FullID: fmt.Sprintf("GOBLIN-%s", gatehouse.ID),
				}
			} else if to != "" {
				var err error
				recipientIdentity, err = agent.ParseAgentID(to)
				if err != nil {
					return fmt.Errorf("invalid recipient: %w", err)
				}
			} else {
				return fmt.Errorf("--to or --to-goblin is required")
			}

			// Use subject or generate default
			if subject == "" {
				subject = "(no subject)"
			}

			// Create message (sender and recipient are actor IDs)
			resp, err := wire.MessageService().CreateMessage(ctx, primary.CreateMessageRequest{
				Sender:    identity.FullID,
				Recipient: recipientIdentity.FullID,
				Subject:   subject,
				Body:      body,
			})
			if err != nil {
				return fmt.Errorf("failed to create message: %w", err)
			}

			fmt.Printf("✓ Message sent: %s\n", resp.MessageID)
			fmt.Printf("  From: %s\n", identity.FullID)
			fmt.Printf("  To: %s\n", recipientIdentity.FullID)
			fmt.Printf("  Subject: %s\n", subject)

			// If --nudge flag is set, also send a real-time nudge to the recipient
			if nudge {
				if err := sendNudgeToAgent(ctx, recipientIdentity, fmt.Sprintf("You have new mail: %s", subject)); err != nil {
					// Don't fail the command if nudge fails - mail was still sent
					fmt.Printf("  ⚠ Nudge failed: %v\n", err)
				} else {
					fmt.Printf("  ✓ Nudge sent\n")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Recipient agent ID (e.g., IMP-BENCH-001, GOBLIN-GATE-003)")
	cmd.Flags().BoolVar(&toGoblin, "to-goblin", false, "Send to workshop's goblin (auto-resolved from context)")
	cmd.Flags().StringVar(&subject, "subject", "", "Message subject")
	cmd.Flags().BoolVar(&nudge, "nudge", false, "Also send real-time nudge to recipient")

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
			ctx := NewContext()

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
  orc mail read MSG-COMM-001-005`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
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
  orc mail conversation IMP-WB-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
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

// sendNudgeToAgent sends a real-time nudge to an agent via tmux
func sendNudgeToAgent(ctx context.Context, identity *agent.AgentIdentity, message string) error {
	tmuxAdapter := wire.TMuxAdapter()

	var target string

	if identity.Type == agent.AgentTypeGoblin {
		if identity.ID == "GOBLIN" {
			// Legacy generic GOBLIN address - check for ORC session
			sessionName := "ORC"
			if !tmuxAdapter.SessionExists(ctx, sessionName) {
				return fmt.Errorf("tmux session %s not running", sessionName)
			}
			target = "ORC:1.1"
		} else {
			// GOBLIN-GATE-XXX: lookup gatehouse → workshop → session
			gatehouse, err := wire.GatehouseService().GetGatehouse(ctx, identity.ID)
			if err != nil {
				return fmt.Errorf("failed to get gatehouse info: %w", err)
			}

			sessionName := tmuxAdapter.FindSessionByWorkshopID(ctx, gatehouse.WorkshopID)
			if sessionName == "" {
				return fmt.Errorf("no tmux session found for workshop %s", gatehouse.WorkshopID)
			}

			// Gatehouse is window 1, pane 1 (Claude)
			target = fmt.Sprintf("%s:1.1", sessionName)
		}
	} else {
		// IMP: lookup workbench and find session by workshop ID
		workbench, err := wire.WorkbenchService().GetWorkbench(ctx, identity.ID)
		if err != nil {
			return fmt.Errorf("failed to get workbench info: %w", err)
		}

		// Find session by workshop ID (runtime lookup)
		sessionName := tmuxAdapter.FindSessionByWorkshopID(ctx, workbench.WorkshopID)
		if sessionName == "" {
			return fmt.Errorf("no tmux session found for workshop %s", workbench.WorkshopID)
		}

		// Window named by workbench, pane 2 is Claude
		target = fmt.Sprintf("%s:%s.2", sessionName, workbench.Name)
	}

	return tmuxAdapter.NudgeSession(ctx, target, message)
}
