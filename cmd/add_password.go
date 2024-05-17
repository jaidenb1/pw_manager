package cmd

import (
    "fmt"
    "os"
	"github.com/spf13/cobra"
    "crypto/aes"
	"crypto/cipher"
	"encoding/base64"

)

var passwordAdd= &cobra.Command{
	Use:   "add",
	Short: "Add a new password",
    //Long: "long",
	// Uncomment the following line if your bare application
	// has an action associated with it:
     Run: addPassword,
}

func init() {
    
    //rootCmd.
    rootCmd.AddCommand(passwordAdd)
}


func addPassword(cmd *cobra.Command, arg []string) {

    //var args []string

    //setMaster(cmd, args)
    login()
    
    var title string
    var username string
    var password string

    fmt.Println("What is the password for? ")
    fmt.Scan(&title)

    fmt.Println("Enter username: ")
    fmt.Scan(&username)

    fmt.Println("Enter password: ")
    fmt.Scan(&password)

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


