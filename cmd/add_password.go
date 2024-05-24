package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var passwordAdd = &cobra.Command{
	Use:   "add",
	Short: "Add a new password",
	// Long: "long",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: addPassword,
}

func init() {
	// rootCmd.
	rootCmd.AddCommand(passwordAdd)
}

var (
	purpleStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = purpleStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).Foreground(lipgloss.Color("205"))
)

type add_model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
}

func addPassword(cmd *cobra.Command, arg []string) {
	login()
	p := tea.NewProgram(initialModelAdd())
	if m, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	} else {
		// fmt.Println(m)
		if model, ok := m.(add_model); ok {

			// fmt.Println(reflect.TypeOf(model.inputs))
			input_map := GetInputValues(model.inputs)

			// check if any entries contain only white space
			for key, value := range input_map {
				if len(strings.TrimSpace(value)) == 0 {
					fmt.Println(borderStyle.Render("Error: invalid " + key))
					os.Exit(1)

				}
			}

			if input_map["Password"] != input_map["Confirm Password"] {
				fmt.Println(borderStyle.Render("Error: passwords do not match"))
			} else {
				writePassword(input_map["Title"], input_map["Username/Email"], input_map["Password"])
				fmt.Println(purpleStyle.Render("Password successfully added"))
			}
		}
	}
}

func initialModelAdd() add_model {
	m := add_model{
		inputs: make([]textinput.Model, 4),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Title"
			t.Focus()
			t.PromptStyle = purpleStyle
			t.TextStyle = purpleStyle
		case 1:
			t.Placeholder = "Username/Email"
			t.CharLimit = 64
		case 2:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'

		case 3:
			t.Placeholder = "Confirm Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m add_model) Init() tea.Cmd {
	return textinput.Blink
}

func (m add_model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *add_model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m add_model) View() string {
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

func GetInputValues(m []textinput.Model) map[string]string {
	// var inputMap map[string]string

	inputMap := make(map[string]string)

	for _, input := range m {
		inputMap[input.Placeholder] = input.Value()

		// fmt.Printf(input.Value())
		// fmt.Printf("%s: %s\n", input.Placeholder, input.Value())

	}
	// fmt.Printf("%s: %s\n", input.Placeholder, input.Value())

	return inputMap
}

func writePassword(title string, username string, password string) {
	encPass, err := Encrypt(password, MySecret)
	if err != nil {
		fmt.Println("error encrypting your classified text: ", err)
	}

	f, err := os.Create("passwords/" + title + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	l, err := f.WriteString(username + "\n" + string(encPass))
	if err != nil {
		fmt.Println(l, err)
		f.Close()
		return
	}
}

var bytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

// This should be in an env file in production
const MySecret string = "a*~c&#s=)^^1b2%^^#70^b34"

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// Encrypt method is to encrypt or hide any classified text
func Encrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)

	return Encode(cipherText), nil
}
