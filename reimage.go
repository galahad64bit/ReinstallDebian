package reimage

import (
  "bytes"
  "context"
  "errors"
  "fmt"
  "os"
  "os/exec"
  "os/user"
  "strings"
  
)

func main(){
  ctx := context.Background()
  
  if err := rootCheck(); err != nil{
    log.ExitContext(ctx, "Failed to run command as root")
  }
  if err := secureBootEnabled(ctx); err != nil {
    log.ExitContext(ctx, "Failed at Secureboot check")
  }
  if err := secondDisk(); err != nil {
    log.ExitContextf(ctx, "Failed at checking for secondary drives")
  }
  if err := ensureNoMultiUser(ctx); err != nil {
		log.ExitContextf(ctx, "Failed at ensureNoMultiUser")
	}
  fmt.Println("Are you sure you want to reinstall the OS?[y/n]:")
  reader := bufio.NewReader(os.Stdin)
  input, err := reader.ReadString('\n')
  if err != nil{
    log.ExitContextf(ctx, "Failed ot get user input")
  }
  input = strings.TrimpSpace(input)
    
  if strings.ToLower(input) == "n" || strings.ToLower(input) == "no" {
      fmt.ExitContextf("exiting due to user input")
  }
  if err := reimage(ctx); err != nil {
    log.ExitContextf(ctx, "Failed to reinstall OS")
  }
  
}
//Check if command is being ran as root 
func rootCheck(){
  u, err := user.Current()
  if err != nil {
    return fmt.Error("unable to get user")
  }
  if u.Username == "root"{
    return nil
  }
  
  fmt.Println("Rerunning with sudo")
  return syscall.Exec("/usr/bin/sudo", append([]string{"/usr/bin/sudo", "--preserve-env", os.Args[0]}, os.Args[1:]...), os.Environ())
}
// Check if secureboot is enabled.
func secureBootEnabled(ctx context.Context) error {
	if output, err := cmd.Output(ctx, "bootctl", cmd.WithArgs("--quiet")); err != nil {
		return fmt.Errorf("failed at running command: bootctl --quiet")
	} else if !strings.Contains(string(output), "Secure Boot: enabled (user)") {
		fmt.Println("\nSecureboot is not set up")
		return fmt.Error("aborting as secureboot is not enabled")
	}
	return nil
}
// Check for multiple disks installed, don't want to accidentally wipe a drive
func secondDisk() error {
        cmd := exec.Command("lsblk")
        output, err := cmd.Output()
        if err != nil {
                return fmt.Errorf("Error checking disks: %v", err)
        }
        num := strings.Count(string(output), "disk")
        if num != 1 {
                fmt.Println("\nMultiple Disks found. Additional drives are not supported by Techstop or gLinux.")
                return fmt.Errorf("please remove additional drives and try again\n")
        }
        return nil
}
// Multi-user homedir check on device
func ensureNoMultiUser(ctx context.Context) error {
	files, err := file.Match(ctx, "/usr/local/home/*", file.StatNone)
	if err != nil {
		return fmt.Errorf("error at checking for multiple home directories: %v", err)
	}
	count := len(files)
	if count == 1 {
		return nil
	}
	fmt.Println("\nUser home directories found on device:")
	for _, file := range files {
		if file.FileInfo().IsDir() {
			fmt.Println(file.FileInfo().Name())
		}
	}
	if count > 1 {
		fmt.Println("Important: This computer is used by multiple people. Reimaging it will permanently delete files for everyone.")
		fmt.Println("Do you want to continue reinstalling the OS?[y/n]: ")
    reader := bufio.NewReader(os.Stdin)
    input, err := reader.Readstring('\n')
    if err != nil {
      fmt.Error("Failed to get user input")
    }
    input = strings.TrimpSpace(input)
    
    if strings.ToLower(input) == "n" || strings.ToLower(input) == "no"{
      return fmt.Error("exiting due to user input")
    }
	}
	return nil
}
// function that changes loader.conf to default to recovery image for reinstallion of OS
func reimage(ctx context.Context) error {
  filepath= "/path/to/recovery/image.conf"
  var b bytes.Buffer
	b.WriteString("timeout 3\n")
	b.WriteString(fmt.Sprintf("default %s\n", filepath ))
	if err := file.WriteFile(ctx, "/boot/efi/loader/loader.conf", b.Bytes()); err != nil {
		return fmt.Errorf("failed to write loader.conf: %v", err)
	}
	output, err := cmd.Output(ctx, "reboot", cmd.WithArgs("now"))
	log.InfoContextf(ctx, "reboot debug: %s", output)
	if err != nil {
		return err
	}

	return nil
}

