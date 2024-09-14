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
	sampleRate = 44100
	duration   = 2 * time.Second
	bitDepth   = 16
	amplitude  = math.MaxInt16 // Maximum amplitude for int16
)

// Waveform function type that represents a wave generator
type Waveform func(t float64, frequency float64, amplitude float64, phase float64) float64

// SineWave generates a sine wave value for a given time (t), frequency, amplitude, and phase
func SineWave(t float64, frequency float64, amplitude float64, phase float64) float64 {
	return amplitude * math.Sin(2.0*math.Pi*frequency*t+phase)
}

// SquareWave generates a square wave value for a given time (t), frequency, amplitude, and phase
func SquareWave(t float64, frequency float64, amplitude float64, phase float64) float64 {
	if math.Sin(2.0*math.Pi*frequency*t+phase) >= 0 {
		return amplitude
	}
	return -amplitude
}

// TriangleWave generates a triangle wave value for a given time (t), frequency, amplitude, and phase
func TriangleWave(t float64, frequency float64, amplitude float64, phase float64) float64 {
	return (2 * amplitude / math.Pi) * math.Asin(math.Sin(2.0*math.Pi*frequency*t+phase))
}

// GenerateWave generates the samples for a given waveform function, frequency, and duration
func GenerateWave(waveFunc Waveform, frequency float64, amplitude float64, phase float64, sampleRate int, duration time.Duration) []int16 {
	numSamples := int(duration.Seconds() * float64(sampleRate))
	wave := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		sample := waveFunc(t, frequency, amplitude, phase)
		wave[i] = int16(sample)
	}

	return wave
}

// PlayWave plays the generated waveform using SDL2_mixer from a temporary file
func PlayWave(wave []int16, sampleRate int) error {
	// Write the wave to a temporary file
	tmpfile, err := ioutil.TempFile("", "waveform_*.wav")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if err := WriteWAV(tmpfile.Name(), wave, sampleRate); err != nil {
		return fmt.Errorf("failed to write wave to file: %v", err)
	}

	// Initialize SDL2 mixer
	if err := mix.OpenAudio(sampleRate, mix.DEFAULT_FORMAT, 1, 4096); err != nil {
		return fmt.Errorf("failed to initialize audio: %v", err)
	}
	defer mix.CloseAudio()

	// Load the temporary wave file
	chunk, err := mix.LoadWAV(tmpfile.Name())
	if err != nil {
		return fmt.Errorf("failed to load WAV: %v", err)
	}
	defer chunk.Free()

	// Play the waveform
	if _, err := chunk.Play(-1, 0); err != nil {
		return fmt.Errorf("failed to play wave: %v", err)
	}

	time.Sleep(duration) // Allow the sound to play for the duration

	return nil
}

// WriteWAV writes the generated waveform to a WAV file
func WriteWAV(filename string, wave []int16, sampleRate int) error {
	buffer := &audio.IntBuffer{
		Data:           make([]int, len(wave)),
		Format:         &audio.Format{SampleRate: sampleRate, NumChannels: 1},
		SourceBitDepth: bitDepth,
	}

	for i, sample := range wave {
		buffer.Data[i] = int(sample)
	}

	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer outFile.Close()

	encoder := wav.NewEncoder(outFile, sampleRate, bitDepth, 1, 1)
	if err := encoder.Write(buffer); err != nil {
		return fmt.Errorf("failed to write WAV data: %v", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close WAV encoder: %v", err)
	}

	fmt.Printf("Written %s\n", filename)
	return nil
}

// CombineWaves combines multiple waveforms into one by summing their values
func CombineWaves(waves ...[]int16) []int16 {
	if len(waves) == 0 {
		return nil
	}

	numSamples := len(waves[0])
	combined := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		sum := int32(0)
		for _, wave := range waves {
			sum += int32(wave[i])
		}
		// Ensure the combined value does not exceed the int16 range
		if sum > math.MaxInt16 {
			sum = math.MaxInt16
		} else if sum < math.MinInt16 {
			sum = math.MinInt16
		}
		combined[i] = int16(sum)
	}

	return combined
}

func main() {
	frequency := 220.0
	amplitude := 0.8 * amplitude // 80% of full amplitude
	phase := 0.0

	fmt.Println("Generating waveforms...")

	// Generate sine wave
	sineWave := GenerateWave(SineWave, frequency, amplitude, phase, sampleRate, duration)
	// Play and then write the sine wave
	fmt.Println("Playing sine wave...")
	if err := PlayWave(sineWave, sampleRate); err != nil {
		log.Fatalf("Error playing sine wave: %v", err)
	}
	if err := WriteWAV("sine_wave.wav", sineWave, sampleRate); err != nil {
		log.Fatalf("Error writing sine_wave.wav: %v", err)
	}

	// Generate square wave
	squareWave := GenerateWave(SquareWave, frequency, amplitude, phase, sampleRate, duration)
	// Play and then write the square wave
	fmt.Println("Playing square wave...")
	if err := PlayWave(squareWave, sampleRate); err != nil {
		log.Fatalf("Error playing square wave: %v", err)
	}
	if err := WriteWAV("square_wave.wav", squareWave, sampleRate); err != nil {
		log.Fatalf("Error writing square_wave.wav: %v", err)
	}

	// Generate triangle wave
	triangleWave := GenerateWave(TriangleWave, frequency, amplitude, phase, sampleRate, duration)
	// Play and then write the triangle wave
	fmt.Println("Playing triangle wave...")
	if err := PlayWave(triangleWave, sampleRate); err != nil {
		log.Fatalf("Error playing triangle wave: %v", err)
	}
	if err := WriteWAV("triangle_wave.wav", triangleWave, sampleRate); err != nil {
		log.Fatalf("Error writing triangle_wave.wav: %v", err)
	}

	// Combine all three waveforms
	fmt.Println("Combining all waveforms...")
	combinedWave := CombineWaves(sineWave, squareWave, triangleWave)

	// Play and write the combined waveform
	fmt.Println("Playing combined waveform...")
	if err := PlayWave(combinedWave, sampleRate); err != nil {
		log.Fatalf("Error playing combined waveform: %v", err)
	}
	if err := WriteWAV("combined_wave.wav", combinedWave, sampleRate); err != nil {
		log.Fatalf("Error writing combined_wave.wav: %v", err)
	}

	fmt.Println("Done!")
}
