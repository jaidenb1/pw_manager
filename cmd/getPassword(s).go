package cmd

import (
    "fmt"
    "os"
	"github.com/spf13/cobra"
    "crypto/aes"
	"crypto/cipher"
	"encoding/base64"
    "strings"
    "bufio"

)

var allPasswords bool

var passwordGet= &cobra.Command{
	Use:   "get",
	Short: "Get passwords",
    //Long: "long",
	// Uncomment the following line if your bare application
	// has an action associated with it:
     Run: getPassword,
}

func init() {
    
    //rootCmd.
    passwordGet.Flags().BoolVarP(&allPasswords, "all", "a", false, "Show all passwords")
    rootCmd.AddCommand(passwordGet)
}


func getPassword(cmd *cobra.Command, arg []string) {

    login()

    if allPasswords {
        showAllPasswords()
    } else {
        getSinglePassword()
    }

}

func getSinglePassword() {

    var selected string

    fmt.Println("Choose a password:")
    fmt.Println()

    all_files := getDirContents()

    for _, file := range all_files {

        split := strings.Split(file, ".")
        fmt.Println(split[0])        
    }

    fmt.Println("")
    fmt.Scan(&selected)

    
    file, err := os.Open("./passwords/" + selected + ".txt")
    if err != nil {
        fmt.Println(err)
    }
    defer file.Close()

    displayUserAndPassword(file)

}


func showAllPasswords() {

    files, err := os.ReadDir("./passwords")
    if err != nil {
        fmt.Println(err)
        return
    }
    for _, file := range files {

        file_name := strings.Split(file.Name(), ".")[0]

        file, err := os.Open("./passwords/" + file.Name())
        if err != nil {
            fmt.Println(err)
        }
        defer file.Close()

        fmt.Println(file_name + ":")
        displayUserAndPassword(file)

    }
}

func displayUserAndPassword(file *os.File)  {

    var lines []string

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

func getDirContents() []string {

    var file_arr []string

    dir, err := os.Open("./passwords")
	if err != nil {
		fmt.Println("Error opening directory:", err)
		return file_arr
	}
	defer dir.Close()

    files, err := dir.Readdirnames(-1)
    if err != nil {
        fmt.Println(err)
        return file_arr
    }

    for _, file := range files {
        file_arr = append(file_arr, file)
    }

    return file_arr

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




