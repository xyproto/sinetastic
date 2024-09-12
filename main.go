package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/veandco/go-sdl2/mix"
)

const (
	frequency1 = 220.0
	frequency2 = 110.0
	duration   = 2 * time.Second
	sampleRate = 44100
	bitDepth   = 16
)

func GenerateSineWave(freq float64, sampleRate int, duration time.Duration) []int16 {
	numSamples := int(float64(sampleRate) * duration.Seconds())
	sineWave := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		amplitude := math.MaxInt16
		sineWave[i] = int16(float64(amplitude) * math.Sin(2.0*math.Pi*freq*t))
	}

	return sineWave
}

func WriteSineWaveToTempFile(wave []int16, sampleRate int) (string, error) {
	tmpfile, err := ioutil.TempFile("", "sine_wave_*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpfile.Close()

	buffer := &audio.IntBuffer{
		Data:           make([]int, len(wave)),
		Format:         &audio.Format{SampleRate: sampleRate, NumChannels: 1},
		SourceBitDepth: bitDepth,
	}

	for i, sample := range wave {
		buffer.Data[i] = int(sample)
	}

	encoder := wav.NewEncoder(tmpfile, sampleRate, bitDepth, 1, 1)
	if err := encoder.Write(buffer); err != nil {
		return "", fmt.Errorf("failed to write WAV data: %v", err)
	}
	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("failed to close WAV encoder: %v", err)
	}

	return tmpfile.Name(), nil
}

func PlaySineWavesSimultaneously(file1, file2 string) error {
	if err := mix.OpenAudio(sampleRate, mix.DEFAULT_FORMAT, 2, 4096); err != nil {
		return fmt.Errorf("failed to open SDL2_mixer audio: %v", err)
	}
	defer mix.CloseAudio()

	chunk1, err := mix.LoadWAV(file1)
	if err != nil {
		return fmt.Errorf("failed to load WAV file1: %v", err)
	}
	defer chunk1.Free()

	var chunk2 *mix.Chunk
	if file2 != "" {
		chunk2, err = mix.LoadWAV(file2)
		if err != nil {
			return fmt.Errorf("failed to load WAV file2: %v", err)
		}
		defer chunk2.Free()
	}

	fmt.Println("Playing first sine wave...")
	if _, err := chunk1.Play(-1, 0); err != nil {
		return fmt.Errorf("failed to play WAV file1: %v", err)
	}

	if chunk2 != nil {
		fmt.Println("Playing second sine wave...")
		if _, err := chunk2.Play(-1, 0); err != nil {
			return fmt.Errorf("failed to play WAV file2: %v", err)
		}
	}

	time.Sleep(duration)

	return nil
}

func main() {
	fmt.Println("Generating sine waves")
	sineWave1 := GenerateSineWave(frequency1, sampleRate, duration)
	sineWave2 := GenerateSineWave(frequency2, sampleRate, duration)

	fmt.Println("Writing sine waves to temporary files")
	tmpfile1, err := WriteSineWaveToTempFile(sineWave1, sampleRate)
	if err != nil {
		log.Fatalf("Error writing 220 Hz sine wave: %v", err)
	}
	defer os.Remove(tmpfile1)

	tmpfile2, err := WriteSineWaveToTempFile(sineWave2, sampleRate)
	if err != nil {
		log.Fatalf("Error writing 110 Hz sine wave: %v", err)
	}
	defer os.Remove(tmpfile2)

	fmt.Println("Playing first sine wave (220 Hz)")
	if err := PlaySineWavesSimultaneously(tmpfile1, ""); err != nil {
		log.Fatalf("Error playing 220 Hz sine wave: %v", err)
	}

	fmt.Println("Playing second sine wave (110 Hz)")
	if err := PlaySineWavesSimultaneously(tmpfile2, ""); err != nil {
		log.Fatalf("Error playing 110 Hz sine wave: %v", err)
	}

	fmt.Println("Playing both sine waves together")
	if err := PlaySineWavesSimultaneously(tmpfile1, tmpfile2); err != nil {
		log.Fatalf("Error playing sine waves simultaneously: %v", err)
	}
}
