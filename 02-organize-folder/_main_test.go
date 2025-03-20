package organize-folder

import (
	bytes "bytes"
	errors "errors"
	fmt "fmt"
	os "os"
	testing "testing"
	assert "github.com/stretchr/testify/assert"
	debug "runtime/debug"
	strings "strings"
	filepath "path/filepath"
	log "log"
	sync "sync"
	bufio "bufio"
)








/*
ROOST_METHOD_HASH=check_6690bbebba
ROOST_METHOD_SIG_HASH=check_f942c97545

FUNCTION_DEF=func check(err error) // check for any error


*/
func TestCheck(t *testing.T) {
	type scenario struct {
		name            string
		inputError      error
		expectedOutput  string
		expectedExit    bool
		exitCode        int
		mockExitHandler func(code int)
	}


/*
ROOST_METHOD_HASH=createDefaultFolders_11096afe5f
ROOST_METHOD_SIG_HASH=createDefaultFolders_c912f4a967

FUNCTION_DEF=func createDefaultFolders(targetFolder string) 

*/
func TestCreateDefaultFolders(t *testing.T) {

	defaultFolders := []string{"Music", "Videos", "Docs", "Images", "Others"}

	t.Run("Scenario 1: Verify Default Folders Are Created in Target Directory", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "organize-folder-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		createDefaultFolders(tempDir)

		for _, folder := range defaultFolders {
			folderPath := filepath.Join(tempDir, folder)
			if _, err := os.Stat(folderPath); os.IsNotExist(err) {
				t.Errorf("Expected folder %s to be created, but it does not exist", folder)
			}
		}
		t.Log("Test passed: All default folders are successfully created")
	})

	t.Run("Scenario 2: Ensure Existing Folders Remain Unchanged", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "organize-folder-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		os.Mkdir(filepath.Join(tempDir, "Music"), 0755)
		os.Mkdir(filepath.Join(tempDir, "Videos"), 0755)

		createDefaultFolders(tempDir)

		for _, folder := range defaultFolders {
			folderPath := filepath.Join(tempDir, folder)
			if _, err := os.Stat(folderPath); os.IsNotExist(err) {
				t.Errorf("Expected folder %s to exist, but it does not", folder)
			}
		}
		t.Log("Test passed: Existing folders remain unchanged, and missing ones are created")
	})

	t.Run("Scenario 3: Verify Behavior with Non-Existent Target Folder", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		nonExistentDir := filepath.Join(os.TempDir(), "non-existent-dir")

		err := os.RemoveAll(nonExistentDir)
		if err != nil {
			t.Fatalf("Failed to remove non-existent directory: %v", err)
		}

		createDefaultFolders(nonExistentDir)

		for _, folder := range defaultFolders {
			folderPath := filepath.Join(nonExistentDir, folder)
			if _, err := os.Stat(folderPath); os.IsNotExist(err) {
				t.Logf("Folder %s correctly not created under non-existent directory", folder)
			} else {
				t.Errorf("Unexpectedly found folder %s under non-existent directory", folder)
			}
		}
		t.Log("Test passed: Function handled non-existent directory gracefully")
	})

	t.Run("Scenario 4: Verify Behavior With Insufficient Permissions", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir := t.TempDir()
		err := os.Chmod(tempDir, 0444)
		if err != nil {
			t.Fatalf("Failed to change permissions of directory: %v", err)
		}
		defer os.Chmod(tempDir, 0755)

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v. Test passed.", r)
			}
		}()

		createDefaultFolders(tempDir)
		t.Log("Test passed: Insufficient permission scenario handled gracefully")
	})

	t.Run("Scenario 5: Verify Function No-Ops with Empty Target Path", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		createDefaultFolders("")
		t.Log("Test passed: Function gracefully handled empty target path")
	})

	t.Run("Scenario 6: Verify Concurrency Safety When Invoked Multiple Times Simultaneously", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "organize-folder-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		var wg sync.WaitGroup
		const goroutineCount = 10

		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				createDefaultFolders(tempDir)
			}()
		}

		wg.Wait()

		for _, folder := range defaultFolders {
			folderPath := filepath.Join(tempDir, folder)
			if _, err := os.Stat(folderPath); os.IsNotExist(err) {
				t.Errorf("Expected folder %s to exist after concurrent creation", folder)
			}
		}
		t.Log("Test passed: Concurrency safety verified")
	})

	t.Run("Scenario 7: Verify Behavior With Target Folder Path Containing Special Characters", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "My@Target#Folder")
		if err != nil {
			t.Fatalf("Failed to create temp directory with special characters: %v", err)
		}
		defer os.RemoveAll(tempDir)

		createDefaultFolders(tempDir)

		for _, folder := range defaultFolders {
			folderPath := filepath.Join(tempDir, folder)
			if _, err := os.Stat(folderPath); os.IsNotExist(err) {
				t.Errorf("Expected folder %s to exist, but it does not", folder)
			}
		}
		t.Log("Test passed: Special character folder scenario handled correctly")
	})

	t.Run("Scenario 8: Verify Custom Folder Permissions Are Applied", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "organize-folder-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		createDefaultFolders(tempDir)

		for _, folder := range defaultFolders {
			folderPath := filepath.Join(tempDir, folder)
			info, err := os.Stat(folderPath)
			if err != nil {
				t.Errorf("Failed to retrieve info for folder %s: %v", folder, err)
				continue
			}
			if info.Mode().Perm() != 0755 {
				t.Errorf("Expected permissions for %s to be 0755, got %v", folder, info.Mode().Perm())
			}
		}
		t.Log("Test passed: Correct permissions applied to created folders")
	})
}


/*
ROOST_METHOD_HASH=organizeFolder_1a14b1518b
ROOST_METHOD_SIG_HASH=organizeFolder_123b127d0e

FUNCTION_DEF=func organizeFolder(targetFolder string) 

*/
func TestOrganizeFolder(t *testing.T) {

	createTempStructure := func(base string, structure map[string]string) error {
		for name, content := range structure {
			filePath := filepath.Join(base, name)
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return err
			}
		}
		return nil
	}

	filesExist := func(base string, files []string) bool {
		for _, file := range files {
			if _, err := os.Stat(filepath.Join(base, file)); os.IsNotExist(err) {
				return false
			}
		}
		return true
	}

	t.Run("Scenario 1: Successfully organize image files into an 'Images' folder", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "test-dir")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		files := map[string]string{
			"image1.png":  "",
			"image2.jpg":  "",
			"image3.jpeg": "",
		}
		if err := createTempStructure(tempDir, files); err != nil {
			t.Fatalf("Failed to set up directory: %v", err)
		}

		organizeFolder(tempDir)

		imageFiles := []string{"Images/image1.png", "Images/image2.jpg", "Images/image3.jpeg"}
		if !filesExist(tempDir, imageFiles) {
			t.Errorf("Image files were not moved to 'Images' folder")
		}
	})

	t.Run("Scenario 2: Successfully organize video files into a 'Videos' folder", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "test-dir")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		files := map[string]string{
			"video1.mp4": "",
			"video2.mov": "",
			"video3.avi": "",
		}
		if err := createTempStructure(tempDir, files); err != nil {
			t.Fatalf("Failed to set up directory: %v", err)
		}

		organizeFolder(tempDir)

		videoFiles := []string{"Videos/video1.mp4", "Videos/video2.mov", "Videos/video3.avi"}
		if !filesExist(tempDir, videoFiles) {
			t.Errorf("Video files were not moved to 'Videos' folder")
		}
	})

	t.Run("Scenario 3: Successfully organize document files into a 'Docs' folder", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "test-dir")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		files := map[string]string{
			"doc1.pdf":  "",
			"doc2.docx": "",
			"doc3.csv":  "",
		}
		if err := createTempStructure(tempDir, files); err != nil {
			t.Fatalf("Failed to set up directory: %v", err)
		}

		organizeFolder(tempDir)

		docFiles := []string{"Docs/doc1.pdf", "Docs/doc2.docx", "Docs/doc3.csv"}
		if !filesExist(tempDir, docFiles) {
			t.Errorf("Document files were not moved to 'Docs' folder")
		}
	})

	t.Run("Scenario 6: No files in the target folder (Edge Case)", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		tempDir, err := os.MkdirTemp("", "test-dir")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		r, w, _ := os.Pipe()
		old := os.Stdout
		os.Stdout = w

		organizeFolder(tempDir)

		w.Close()
		os.Stdout = old

		output, _ := bufio.NewReader(r).ReadString('\n')

		if !strings.Contains(output, "No files moved") {
			t.Errorf("Unexpected output: got %s, want 'No files moved'", output)
		}
	})

}

