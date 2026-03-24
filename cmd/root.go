package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/blackironj/panorama/conv"
)

const (
	defaultEdgeLen     = 1024
	defaultJpegQuality = 75
	minJpegQuality     = 1
	maxJpegQuality     = 100
	minEdgeLen         = 1
)

var validSides = conv.ValidSides()

var (
	inFilePath    string
	outFileDir    string
	inDirPath     string
	edgeLen       int
	sides         []string
	quality       int
	interpolation string
)

var rootCmd = &cobra.Command{
	Use:   "panorama",
	Short: "convert equirectangular panorama img to Cubemap img",
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&inFilePath, "in", "i", "", "input image file path (required if --indir is not specified)")
	rootCmd.Flags().StringVarP(&inDirPath, "indir", "d", "", "input directory path (required if --in is not specified)")
	rootCmd.Flags().StringVarP(&outFileDir, "out", "o", ".", "output file directory path")
	rootCmd.Flags().IntVarP(&edgeLen, "len", "l", defaultEdgeLen, "edge length of a cube face")
	rootCmd.Flags().StringSliceVarP(&sides, "sides", "s", nil, "array of sides [front,back,left,right,top,bottom] (default: all sides)")
	rootCmd.Flags().IntVarP(&quality, "quality", "q", defaultJpegQuality, "jpeg file output quality ranges from 1 to 100 inclusive, higher is better")
	rootCmd.Flags().StringVarP(&interpolation, "interpolation", "p", "bilinear", "interpolation method: nearest, bilinear, bicubic")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(_ *cobra.Command, _ []string) {
	if inFilePath == "" && inDirPath == "" {
		exitWithError("need an input image file path or input directory")
	}
	if inFilePath != "" && inDirPath != "" {
		exitWithError("need only one path, not both")
	}

	if edgeLen < minEdgeLen {
		exitWithError(fmt.Sprintf("edge length must be at least %d", minEdgeLen))
	}
	if quality < minJpegQuality || quality > maxJpegQuality {
		exitWithError(fmt.Sprintf("quality must be between %d and %d", minJpegQuality, maxJpegQuality))
	}

	targetSides, err := resolveTargetSides(sides)
	if err != nil {
		exitWithError(err)
	}

	interp, err := conv.ParseInterpolation(interpolation)
	if err != nil {
		exitWithError(err)
	}

	startTime := time.Now()
	fmt.Println("Start conversion.")

	if inFilePath != "" {
		if err := processSingleImage(inFilePath, outFileDir, targetSides, interp, false); err != nil {
			exitWithError(err)
		}
	} else {
		processDirectory(inDirPath, outFileDir, targetSides, interp)
	}

	elapsed := time.Since(startTime).Seconds()
	fmt.Printf("Processing complete. elapsed: %.2f sec\n", elapsed)
}

func resolveTargetSides(sides []string) ([]string, error) {
	if len(sides) == 0 {
		return validSides, nil
	}
	for _, side := range sides {
		if !isValidSide(side) {
			return nil, fmt.Errorf("invalid side specified: %s, valid sides are %v", side, validSides)
		}
	}
	return sides, nil
}

func processSingleImage(inPath, outDir string, targetSides []string, interp conv.Interpolator, needSubdir bool) error {
	inImage, ext, err := conv.ReadImage(inPath)
	if err != nil {
		return fmt.Errorf("reading image %s: %w", inPath, err)
	}

	canvases, err := conv.ConvertEquirectangularToCubeMap(edgeLen, inImage, targetSides, interp)
	if err != nil {
		return fmt.Errorf("converting image %s: %w", inPath, err)
	}

	if needSubdir {
		outDir = filepath.Join(outDir, strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath)))
	}

	if err := conv.WriteImage(canvases, outDir, ext, targetSides, quality); err != nil {
		return fmt.Errorf("writing images for %s: %w", inPath, err)
	}

	return nil
}

func processDirectory(inDir, outDir string, targetSides []string, interp conv.Interpolator) {
	files, err := os.ReadDir(inDir)
	if err != nil {
		exitWithError(err)
	}

	var imageFiles []fs.DirEntry
	for _, file := range files {
		if !file.IsDir() && isImageFile(file) {
			imageFiles = append(imageFiles, file)
		}
	}

	totalFiles := len(imageFiles)
	if totalFiles == 0 {
		fmt.Println("No image files found in directory.")
		return
	}

	maxConcurrent := min(runtime.NumCPU(), totalFiles)

	var (
		mu             sync.Mutex
		processedFiles int
		errors         []string
		startTime      = time.Now()
		semaphore      = make(chan struct{}, maxConcurrent)
	)

	var wg sync.WaitGroup
	for _, file := range imageFiles {
		wg.Add(1)
		go func(file fs.DirEntry) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			inPath := filepath.Join(inDir, file.Name())
			if err := processSingleImage(inPath, outDir, targetSides, interp, true); err != nil {
				mu.Lock()
				errors = append(errors, err.Error())
				processedFiles++
				mu.Unlock()
				return
			}

			mu.Lock()
			processedFiles++
			mu.Unlock()
		}(file)
	}

	// Progress reporting
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			time.Sleep(1 * time.Second)
			mu.Lock()
			processed := processedFiles
			mu.Unlock()

			printProgress(processed, totalFiles, startTime)

			if processed >= totalFiles {
				return
			}
		}
	}()

	wg.Wait()
	<-done

	if len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "\nErrors:")
		for _, e := range errors {
			fmt.Fprintln(os.Stderr, e)
		}
	}
}

func printProgress(processed, total int, startTime time.Time) {
	elapsed := time.Since(startTime).Seconds()
	if processed == 0 {
		fmt.Fprintf(os.Stderr, "\rProgress: 0/%d files processed. Elapsed: %.2f seconds", total, elapsed)
		return
	}
	remaining := total - processed
	eta := float64(remaining) / (float64(processed) / elapsed)
	ips := float64(processed) / elapsed
	fmt.Fprintf(os.Stderr, "\rProgress: %d/%d files processed. ETA: %.2f seconds. IT/S: %.2f", processed, total, eta, ips)
	if processed >= total {
		fmt.Fprintln(os.Stderr)
	}
}

func isImageFile(file fs.DirEntry) bool {
	ext := strings.ToLower(filepath.Ext(file.Name()))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func isValidSide(side string) bool {
	return slices.Contains(validSides, side)
}

func exitWithError(msg any) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}
