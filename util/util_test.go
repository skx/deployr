package util

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// TestHash tests the hash of known-contents are correct.
func TestHash(t *testing.T) {

	//
	// The string we'll hash
	//
	input := []byte("This is a test string\n")

	//
	// Generate a temporary file
	//
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	//
	// Write out the data to the file.
	//
	ioutil.WriteFile(tmpfile.Name(), input, 0644)

	//
	// Cleanup when we're done.
	//
	defer os.Remove(tmpfile.Name()) // clean up

	//
	// Now hash the contents.
	//
	content, err := HashFile(tmpfile.Name())

	//
	// We shouldn't see an error.
	//
	if err != nil {
		t.Errorf("We received an unexpected error: %s\n", err.Error())
	}

	//
	// Did we get what we expect?
	//
	expected := "c4af1fb1a2c5a9b67305424eda71bcd91e183f2c"
	if content != expected {
		t.Errorf("Hash failed - expected:%s received:%s",
			expected, content)
	}
}

// TestHashMissing tests that hashing a missing file fails appropriately.
func TestHashMissing(t *testing.T) {

	//
	// Hash the contents of a missing-file
	//
	_, err := HashFile("/not/present/file.CON$!")

	//
	// We shouldn't see an error.
	//
	if err == nil {
		t.Errorf("We expected an error, but received none.\n")
	}
}

// TestFileExists tests our file-testing function.
func TestFileExists(t *testing.T) {

	//
	// Generate a temporary file
	//
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	//
	// Ensure it exists.
	//
	if !FileExists(tmpfile.Name()) {
		t.Errorf("Testing a present-file failed")
	}

	//
	// Now remove it and test it is gone.
	//
	os.Remove(tmpfile.Name())

	if FileExists(tmpfile.Name()) {
		t.Errorf("Missing file is still present?")
	}
}
