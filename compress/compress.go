package compress

import (
    "compress/gzip"
    "io"
    "os"
    "strings"
)

// CompressFile compresses the given filePath and creates a .zip file in the same directory.
func CompressFile(filePath string) error {
   // Open source file
    inFile, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer inFile.Close()

    // Create destination file
    gzPath := filePath + ".gz"
    if strings.HasSuffix(strings.ToLower(filePath), ".gz") {
        gzPath = filePath // Prevent double .gz
    }
    outFile, err := os.Create(gzPath)
    if err != nil {
        return err
    }
    defer outFile.Close()

    // Create a gzip writer on the output file
    gzWriter, err := gzip.NewWriterLevel(outFile, 6) // TODO: Parameterize the compression level arg
    if err != nil {
        return err
    }
    defer gzWriter.Close()

    // Optionally set the original filename in the gzip header
    gzWriter.Name = filePath

    // Copy input file to gzip writer (compresses data)
    if _, err = io.Copy(gzWriter, inFile); err != nil {
        return err
    }

    return nil
}