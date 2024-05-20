package cmd

import (
    "fmt"
    "os"
    "strings"
    "bufio"
    "crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/spf13/cobra"
    tea "github.com/charmbracelet/bubbletea"
    //"golang.org/x/term"


)

var bubbleTea= &cobra.Command{
	Use:   "get",
	Short: "Get passwords",
    //Long: "long",
	// Uncomment the following line if your bare application
	// has an action associated with it:
     Run: bubbleTeaRun,
}

func init() {
    
    //rootCmd.
    //passwordGet.Flags().BoolVarP(&allPasswords, "all", "a", false, "Show all passwords")
    rootCmd.AddCommand(bubbleTea)
}

type Selection struct {
	Choices []string
}

func (s *Selection) Update(value string) {
	s.Choices = append(s.Choices, value)
}

type model struct {
    choices  []string           // items on the to-do list
    cursor   int                // which to-do list item our cursor is pointing at
    selected map[int]struct{}   // which to-do items are selected
    choice   *Selection
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
            fmt.Println("\nSelected passwords:\n ")
            for _, choice := range model.choice.Choices {

                fmt.Println(choice + ":")

                displayUserAndPassword(choice)
            }
        }
    }
}


func initialModel() model {

    password_list := getPasswords()
    
	return model{
		// Our to-do list is a grocery list


		//choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
        choices: password_list,

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
        choice: &Selection{},

	}
}

func (m model) Init() tea.Cmd {
    // Just return `nil`, which means "no I/O right now, please."
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    // Is it a key press?
    case tea.KeyMsg:

        // Cool, what was the actual key pressed?
        switch msg.String() {

        // These keys should exit the program.
        case "ctrl+c", "q":
            return m, tea.Quit

        // The "up" and "k" keys move the cursor up
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }

        // The "down" and "j" keys move the cursor down
        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }

        // The "enter" key and the spacebar (a literal space) toggle
        // the selected state for the item that the cursor is pointing at.
        case "enter", " ":
            _, ok := m.selected[m.cursor]
            if ok {
                delete(m.selected, m.cursor)
            } else {
                m.selected[m.cursor] = struct{}{}
                //m.choice.Update(m.choices[m.cursor])

            }

        case "y":
            for selectedKey := range m.selected {
                //m.choice.Update(m.choices[selectedKey])
                m.cursor = selectedKey
                m.choice.Update(m.choices[m.cursor])
            }
            return m, tea.Quit
		}
    }

    // Return the updated model to the Bubble Tea runtime for processing.
    // Note that we're not returning a command.
    return m, nil
}


func (m model) View() string {
    // The header
    s := "Which password(s) would you like?\n\n"

    // Iterate over our choices
    for i, choice := range m.choices {

        // Is the cursor pointing at this choice?
        cursor := " " // no cursor
        if m.cursor == i {
            cursor = ">" // cursor!
        }

        // Is this choice selected?
        checked := " " // not selected
        if _, ok := m.selected[i]; ok {
            checked = "x" // selected!
        }

        // Render the row
        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    }

    // The footer
    s += "\nPress q to quit.\nPress y to confirm.\n"
    //s += fmt.Sprintf("Press y to confirm choice.\n\n")

    // Send the UI for rendering
    return s
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


func displayUserAndPassword(file_name string)  {

    var lines []string

    file, err := os.Open("./passwords/" + file_name + ".txt")
    if err != nil {
        fmt.Println(err)
    }

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        //fmt.Println(scanner.Text())
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














