package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

const logo = `
 _____       _              ____                                     _     
| ____|_ __ | |_ ___ _ __  |  _ \ __ _ ___ _____      _____  _ __ __| |  _ 
|  _| | '_ \| __/ _ \ '__| | |_) / _' / __/ __\ \ /\ / / _ \| '__/ _' | (_)
| |___| | | | ||  __/ |    |  __/ (_| \__ \__ \\ V  V / (_) | | | (_| |  _ 
|_____|_| |_|\__\___|_|    |_|   \__,_|___/___/ \_/\_/ \___/|_|  \__,_| (_)

`

var setupMaster = &cobra.Command{
	Use:   "set",
	Short: "Set a master password",
	// Long: "long",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: setMaster,
}

func init() {
	rootCmd.AddCommand(setupMaster)
}

func setMaster(cmd *cobra.Command, arg []string) {
	if _, err := os.Stat("masterPW.txt"); err == nil {
		for true {
			fmt.Println(purpleStyle.Render(logo))

			master_in, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return
			}
			// reset password
			if verifyMaster(string(master_in)) {
				startBubbleTeaSetup()
			} else {
				fmt.Println("Incorrect password. Try again.")
			}
		}
	} else {
		startBubbleTeaSetup()
	}
}

func startBubbleTeaSetup() {
	p := tea.NewProgram(initialModelSetup())
	if m, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	} else {
		if model, ok := m.(setup_model); ok {

			input_map := GetInputValues(model.inputs)

			if input_map["Master Password"] != input_map["Confirm Password"] {
				fmt.Println(borderStyle.Render("Error: passwords do not match"))
			} else {
				writeMaster(input_map["Master Password"])
				os.Exit(1)
			}
		}
	}
}

type setup_model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
}

func initialModelSetup() setup_model {
	m := setup_model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Master Password"
			t.Focus()
			t.PromptStyle = purpleStyle
			t.TextStyle = purpleStyle
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 1:
			t.Placeholder = "Confirm Password"
			t.CharLimit = 64
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m setup_model) Init() tea.Cmd {
	return textinput.Blink
}

func (m setup_model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = purpleStyle
					m.inputs[i].TextStyle = purpleStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *setup_model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m setup_model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}

func writeMaster(password string) {
	passwordBytes := []byte(password)

	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	f, err := os.Create("masterPW.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	l, err := f.WriteString(string(hashedPassword))
	if err != nil {
		fmt.Println(l, err)
		f.Close()
		return
	}

	fmt.Println(purpleStyle.Render("Master password saved"))
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func verifyMaster(master_in string) bool {
	content, err := os.ReadFile("masterPW.txt")
	if err != nil {
		// Handle the error
	}

	err = bcrypt.CompareHashAndPassword(content, []byte(master_in))
	// fmt.Println(err) // nil means it is a match

	if err == nil {
		return true
	}

	return false
}

func login() {
	if _, err := os.Stat("masterPW.txt"); err == nil {
		for true {

			fmt.Println(purpleStyle.Render(logo))

			master_in, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return
			}

			if verifyMaster(string(master_in)) {
				break
			} else {
				fmt.Println("Incorrect password. Try again.")
			}
		}
	} else {
	}
}
