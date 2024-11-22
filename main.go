package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	steps    []string
	selected map[int]bool
	cursor   int
	proceed  bool
}

// Initialize the model with the setup steps
func initialModel() model {
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
		selected: make(map[int]bool),
	}
}

// Update function to handle user input
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			// Toggle step selection
			m.selected[m.cursor] = !m.selected[m.cursor]
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
	commands := map[string]string{
		"update_packages":  "apt update && apt upgrade -y && apt autoclean && apt autoremove -y",
		"setup_firewall":   "apt install -y ufw && ufw allow 'OpenSSH' && ufw enable",
		"install_apache":   "apt install -y apache2 curl && systemctl enable apache2 && ufw allow in 'Apache Full' && a2dissite 000-default.conf",
		"install_mariadb":  `apt install -y mariadb-server && mysql_secure_installation`,
		"install_php":      `add-apt-repository ppa:ondrej/php && apt update &&
apt install -y php8.3 libapache2-mod-php php8.3-mysql php8.3-gd php8.3-curl php8.3-xml composer`,
		"install_nodejs":   "apt install -y nodejs npm && npm install -g n && n 20",
		"create_app_user":  `useradd app && usermod -aG sudo,www-data app`,
		"setup_ssh_keys":   "", // Placeholder; SSH setup logic is handled dynamically
		"clone_repository": "", // Placeholder; cloning logic is handled dynamically
		"configure_apache": "", // Placeholder; virtual host logic is handled dynamically
		"restart_apache":   "systemctl restart apache2",
		"reboot_server":    "reboot",
	}

	reader := bufio.NewReader(os.Stdin)

	// Map the steps to corresponding command keys
	stepToCommand := map[int]string{
		0:  "update_packages",
		1:  "setup_firewall",
		2:  "install_apache",
		3:  "install_mariadb",
		4:  "install_php",
		5:  "install_nodejs",
		6:  "create_app_user",
		7:  "setup_ssh_keys",
		8:  "clone_repository",
		9:  "configure_apache",
		10: "restart_apache",
		11: "reboot_server",
	}

	for i, stepSelected := range m.selected {
		if stepSelected {
			commandKey := stepToCommand[i]
			fmt.Printf("Running: %s...\n", m.steps[i])

			// Special handling for dynamic commands
			switch commandKey {
			case "setup_ssh_keys":
				fmt.Println("Paste public SSH keys to add to /home/app/.ssh/authorized_keys (leave empty to finish):")
				var keys []string
				for {
					fmt.Print("> ")
					key, _ := reader.ReadString('\n')
					key = strings.TrimSpace(key)
					if key == "" {
						break
					}
					keys = append(keys, key)
				}

				// Create SSH directory and authorized_keys file with the collected keys
				setupSSHDir := `
mkdir -p /home/app/.ssh &&
chmod 700 /home/app/.ssh &&
touch /home/app/.ssh/authorized_keys &&
chmod 600 /home/app/.ssh/authorized_keys &&
chown -R app:app /home/app/.ssh`
				runCommand(setupSSHDir)

				if len(keys) > 0 {
					writeKeysCmd := fmt.Sprintf(`echo "%s" > /home/app/.ssh/authorized_keys`, strings.Join(keys, "\n"))
					runCommand(writeKeysCmd)
				}

				// Generate SSH key for the app user
				fmt.Println("Generating SSH key for app user...")
				generateSSHKeyCmd := `
su - app -c "ssh-keygen -t rsa -b 4096 -f /home/app/.ssh/id_rsa -q -N ''"`
				runCommand(generateSSHKeyCmd)

				// Read the public key to display it
				publicKeyCmd := `cat /home/app/.ssh/id_rsa.pub`
				fmt.Println("Copy the following SSH public key to your GitHub settings (https://github.com/settings/keys):")
				runCommand(publicKeyCmd)

				// Wait for the user to confirm
				fmt.Println("\nOnce the key is added to GitHub, press Enter to continue.")
				_, _ = reader.ReadString('\n') // Wait for user input

			case "clone_repository":
				fmt.Print("Enter Project Name: ")
				projectName, _ := reader.ReadString('\n')
				projectName = strings.TrimSpace(projectName)

				fmt.Print("Enter Git URL: ")
				gitURL, _ := reader.ReadString('\n')
				gitURL = strings.TrimSpace(gitURL)

				cloneCmd := fmt.Sprintf(`git clone %s /var/www/%s && chown -R www-data:www-data /var/www/%s`,
					gitURL, projectName, projectName)
				runCommand(cloneCmd)

			case "configure_apache":
				fmt.Print("Enter Project Name: ")
				projectName, _ := reader.ReadString('\n')
				projectName = strings.TrimSpace(projectName)

				configureApacheCmd := fmt.Sprintf(`echo '<VirtualHost *:80>
DocumentRoot /var/www/%s/public
<Directory /var/www/%s/public>
AllowOverride All
Options +Indexes
</Directory>
</VirtualHost>' | tee /etc/apache2/sites-available/%s.conf &&
a2ensite %s.conf`, projectName, projectName, projectName, projectName)
				runCommand(configureApacheCmd)

			default:
				// Run the static command
				runCommand(commands[commandKey])
			}
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
