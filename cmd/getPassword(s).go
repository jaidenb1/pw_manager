package cmd

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	focusedStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Bold(true)
	titleStyle            = lipgloss.NewStyle().Background(lipgloss.Color("#01FAC6")).Foreground(lipgloss.Color("#030303")).Bold(true).Padding(0, 1, 0)
	selectedItemStyle     = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("170")).Bold(true)
	selectedItemDescStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("170"))
	descriptionStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#40BDA3"))
)

var bubbleTea = &cobra.Command{
	Use:   "get",
	Short: "Get passwords",
	// Long: "long",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: bubbleTeaRun,
}

func init() {
	// rootCmd.
	// passwordGet.Flags().BoolVarP(&allPasswords, "all", "a", false, "Show all passwords")
	rootCmd.AddCommand(bubbleTea)
}

type Selection struct {
	Choices []string
}

func (s *Selection) Update(value string) {
	s.Choices = append(s.Choices, value)
}

type model struct {
	choices  []string         // items on the to-do list
	cursor          int              // which to-do list item our cursor is pointing at
	selected        map[int]struct{} // which to-do items are selected
	choice          *Selection
	header          string
    filteredChoices []string
    searchQuery     string
}

func bubbleTeaRun(cmd *cobra.Command, arg []string) {
	login()

	p := tea.NewProgram(initialModel())
	if m, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	} else {
		if model, ok := m.(model); ok {
			// Print selected choices after the program exits
			fmt.Println("Selected password(s):\n ")
			for _, choice := range model.choice.Choices {

				fmt.Println(choice + ":")

				displayUserAndPassword(choice)
			}
		}
	}
}

func initialModel() model {
	password_list := getPasswords()
    header := "Select a password:"

	return model{
		// Our to-do list is a grocery list

		// choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		choices: password_list,

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
		choice:   &Selection{},
		header:   purpleStyle.Render(header),
        filteredChoices: password_list, 
        searchQuery: "",
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// possbile feature - allow use of mouse to select passwords (doesnt work rn)
	// case tea.MouseMsg:
	// return m, tea.Printf("(X: %d, Y: %d) %s", msg.X, msg.Y, tea.MouseEvent(msg))

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "backspace": // make backspace work correctly when searching
            if len(m.searchQuery) > 0 {
                m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
                m.filteredChoices = filterChoices(m.choices, m.searchQuery)
            }

		case "down":
			if m.cursor < len(m.filteredChoices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
				// m.choice.Update(m.choices[m.cursor])
			}

		case "1":
			for selectedKey := range m.selected {
				// m.choice.Update(m.choices[selectedKey])
				m.cursor = selectedKey
				m.choice.Update(m.filteredChoices[m.cursor])
			}
			return m, tea.Quit

        default:
            m.cursor = 0 // set cursor to top when searching
            m.searchQuery += msg.String()
            m.filteredChoices = filterChoices(m.choices, m.searchQuery)
            //m.cursor = 0
        }
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {

	s := m.header + "\n\n"
    //s += fmt.Sprintf("Search: %s\n\n", m.searchQuery)
    s += fmt.Sprintf("Search: %s\n\n", purpleStyle.Render(m.searchQuery))
    //s += focusedStyle.Render("Search: " + m.searchQuery + "\n")
    //s = ("Search: " + m.searchQuery + "\n\n")

	// Iterate over our choices
	for i, choice := range m.filteredChoices{

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			// cursor = ">" // cursor!
			cursor = focusedStyle.Render(">")
			choice = selectedItemStyle.Render(choice)
			// choice.Desc = selectedItemDescStyle.Render(choice.Desc)
		} else {
			choice = focusedStyle.Render(choice)
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			// checked = "x" // selected!
			checked = focusedStyle.Render("x")
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	// s += "\nPress q to quit.\nPress y to confirm.\n"
	s += fmt.Sprintf("\nPress %s to confirm.\n\n", focusedStyle.Render("1"))

	// Send the UI for rendering
	return s
}

// function for filtering on search input (query)
func filterChoices(choices []string, query string) []string {
    var filtered []string
    for _, choice := range choices {
        if strings.Contains(strings.ToLower(choice), strings.ToLower(query)) {
            filtered = append(filtered, choice)
        }
    }
    return filtered
}

func getPasswords() []string {
	var password_list []string

	files, err := os.ReadDir("./passwords")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	for _, file := range files {

		file_name := strings.Split(file.Name(), ".")[0]

		password_list = append(password_list, file_name)

	}
	return password_list
}

func displayUserAndPassword(file_name string) {
	var lines []string

	file, err := os.Open("./passwords/" + file_name + ".txt")
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	username := lines[0]
	encPass := lines[1]

	plain_text_pw, err := Decrypt(encPass, MySecret)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Username: " + username)
	fmt.Println("Password: " + plain_text_pw + "\n")
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// Decrypt method is to extract back the encrypted text
func Decrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)

	return string(plainText), nil
}
