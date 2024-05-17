package cmd

import (
    "fmt"
    "os"
	"github.com/spf13/cobra"
    "golang.org/x/crypto/bcrypt"
    "golang.org/x/term"
)

var setupMaster= &cobra.Command{
	Use:   "set",
	Short: "Set a master password",
    //Long: "long",
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
            fmt.Println("Enter master password: ")
            fmt.Println("-----------------------")
            //fmt.Scan(&master_in)

            master_in, err := term.ReadPassword(int(os.Stdin.Fd()))
            if err != nil {
                return
            }
            // reset password
            if verifyMaster(string(master_in)) {
                writeMaster()
                break

            } else {
                fmt.Println("Incorrect password. Try again.")
            }
        }

    } else {
        writeMaster()
    }
}

func writeMaster()  {

    fmt.Println("Enter new master password: ")
    new_master, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return
    }

    fmt.Println("Confirm Password: ")
    new_master_confirm, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return
    }

    if string(new_master_confirm) == string(new_master) {

        hashedPassword, err := bcrypt.GenerateFromPassword(new_master, bcrypt.DefaultCost)
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

        fmt.Println("Master password saved")
        err = f.Close()
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}


func verifyMaster(master_in string) bool {

    content, err := os.ReadFile("masterPW.txt")
    if err != nil {
    // Handle the error
    }

    err = bcrypt.CompareHashAndPassword(content, []byte(master_in))
    //fmt.Println(err) // nil means it is a match

    if err == nil {
        return true
    } 

    return false
}

func login() {

    if _, err := os.Stat("masterPW.txt"); err == nil {

        for true {
            fmt.Println("Enter master password: ")
            fmt.Println("-----------------------")

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





