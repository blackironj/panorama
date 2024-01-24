package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/blackironj/panorama/conv"
)

const defaultEdgeLen = 1024

var (
	inFilePath string
	outFileDir string
	edgeLen    int

	rootCmd = &cobra.Command{
		Use:   "panorama",
		Short: "convert equirectangular panorama img to Cubemap img",
		Run: func(cmd *cobra.Command, args []string) {
			if inFilePath == "" {
				er("Need a image for converting")
			}

			fmt.Println("Read a image...")
			inImage, ext, err := conv.ReadImage(inFilePath)
			if err != nil {
				er(err)
			}

			s := spinner.New(spinner.CharSets[33], 100*time.Millisecond)
			s.FinalMSG = "Complete converting!\n"
			s.Prefix = "Converting..."

			s.Start()
			canvases := conv.ConvertEquirectangularToCubeMap(edgeLen, inImage)
			s.Stop()

			fmt.Println("Write images...")
			if err := conv.WriteImage(canvases, outFileDir, ext); err != nil {
				er(err)
			}
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&inFilePath, "in", "i", "", "in image file path (required)")
	rootCmd.Flags().StringVarP(&outFileDir, "out", "o", ".", "out file dir path")
	rootCmd.Flags().IntVarP(&edgeLen, "len", "l", defaultEdgeLen, "edge length of a cube face")
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		er(err)
	}
}
