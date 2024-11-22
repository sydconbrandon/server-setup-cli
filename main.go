package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	steps       []string
	selected    map[int]bool
	cursor      int
	proceed     bool
	textInput   textinput.Model // For collecting project name and Git URL
	stepInput   int             // Track if we are in the input step
	inputPrompt string          // Current prompt
	inputData   map[string]string
}

// Initialize the model with the setup steps
func initialModel() model {
	input := textinput.New()
	input.Placeholder = "Enter value"
	input.Focus()
	input.CharLimit = 100

	return model{
		steps: []string{
			"Update server packages and dependencies",
			"Install and enable the Firewall",
			"Install Apache2 and cURL",
			"Install MariaDB and configure DB",
			"Install PHP 8.3 and extensions",
			"Install Node.js and npm",
			"Create the app user and set permissions",
			"Set up SSH and GitHub key",
			"Clone project repository",
			"Configure Apache virtual host",
			"Restart Apache",
			"Reboot server",
		},
		selected:    make(map[int]bool),
		textInput:   input,
		stepInput:   -1, // -1 indicates no active input
		inputPrompt: "",
		inputData:   make(map[string]string), // Store input data (e.g., project name and URL)
	}
}

// Update function to handle user input
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.stepInput != -1 {
			// Handle text input for project name or Git URL
			switch msg.String() {
			case "enter":
				// Save input and move to the next prompt or back to the main flow
				if m.stepInput == 0 {
					m.inputData["project_name"] = m.textInput.Value()
					m.stepInput = 1
					m.inputPrompt = "Enter Git URL"
					m.textInput.SetValue("")
					return m, nil
				}
				if m.stepInput == 1 {
					m.inputData["git_url"] = m.textInput.Value()
					m.stepInput = -1 // End input phase
					m.textInput.Blur()
					return m, nil
				}
			case "esc":
				m.stepInput = -1 // Cancel input
				m.textInput.Blur()
				return m, nil
			}

			// Pass input to the textinput model
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// Handle regular navigation and actions
		switch msg.String() {
		case "down", "j":
			if m.cursor < len(m.steps)-1 {
				m.cursor++
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.cursor == 8 {
				// Start the project name and Git URL input
				m.stepInput = 0
				m.inputPrompt = "Enter Project Name"
				m.textInput.SetValue("")
				m.textInput.Focus()
			} else {
				// Toggle step selection
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "q", "esc":
			// Quit without executing commands
			return m, tea.Quit
		case "c":
			// Set proceed to true and quit the interface
			m.proceed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// Render the UI
func (m model) View() string {
	var builder strings.Builder

	if m.stepInput != -1 {
		// Render the input prompt
		builder.WriteString(fmt.Sprintf("%s:\n", m.inputPrompt))
		builder.WriteString(m.textInput.View())
		builder.WriteString("\n\nPress 'Enter' to confirm or 'Esc' to cancel.\n")
		return builder.String()
	}

	// Render the main checklist
	builder.WriteString("Server Setup Checklist\n\n")
	for i, step := range m.steps {
		cursor := "  "
		if i == m.cursor {
			cursor = "👉"
		}
		selected := " "
		if m.selected[i] {
			selected = "✗"
		}
		builder.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, selected, step))
	}
	builder.WriteString("\nPress 'Enter' to select, 'c' to continue, or 'q' to quit without executing.\n")
	return builder.String()
}

// Execute the selected setup steps
func executeSetupSteps(m model) {
	commands := map[int]string{
		0: "apt update && apt upgrade -y && apt autoclean && apt autoremove -y",
		1: "apt install -y ufw && ufw allow 'OpenSSH' && ufw enable",
		2: "apt install -y apache2 curl && systemctl enable apache2 && ufw allow in 'Apache Full' && a2dissite 000-default.conf",
		3: `apt install -y mariadb-server && mysql_secure_installation`,
		4: `add-apt-repository ppa:ondrej/php && apt update &&
apt install -y php8.3 libapache2-mod-php php8.3-mysql php8.3-gd php8.3-curl php8.3-xml composer`,
		5: "apt install -y nodejs npm && npm install -g n && n 20",
		6: `useradd app && usermod -aG sudo,www-data app`,
		7: `mkdir -p /home/app/.ssh && cp /home/root/.ssh/authorized_keys /home/app/.ssh/authorized_keys && chown -R app:app /home/app/.ssh`,
		8: fmt.Sprintf(`git clone %s /var/www/%s && chown -R www-data:www-data /var/www/%s`,
			m.inputData["git_url"], m.inputData["project_name"], m.inputData["project_name"]),
		9: fmt.Sprintf(`echo '<VirtualHost *:80>
DocumentRoot /var/www/%s/public
<Directory /var/www/%s/public>
AllowOverride All
Options +Indexes
</Directory>
</VirtualHost>' | tee /etc/apache2/sites-available/%s.conf &&
a2ensite %s.conf`, m.inputData["project_name"], m.inputData["project_name"], m.inputData["project_name"], m.inputData["project_name"]),
		10: "systemctl restart apache2",
		11: "reboot",
	}

	for i, stepSelected := range m.selected {
		if stepSelected {
			fmt.Printf("Running: %s...\n", m.steps[i])
			runCommand(commands[i])
		}
	}
}

// Helper function to execute shell commands
func runCommand(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return
	}

	m, ok := finalModel.(model)
	if !ok {
		fmt.Println("Error: unable to cast final model to custom type")
		return
	}

	if m.proceed {
		fmt.Println("Executing selected setup steps...")
		executeSetupSteps(m)
	} else {
		fmt.Println("Aborted. No steps were executed.")
	}
}
