package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	frequency1   = 220.0
	frequency2   = 110.0
	duration     = 2 * time.Second
	sampleRate44 = 44100
	sampleRate48 = 48000
	bitDepth     = 16
	waveform     = "sine"
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

func WriteWaveToFile(filename string, wave []int16, sampleRate int) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	buffer := &audio.IntBuffer{
		Data:           make([]int, len(wave)),
		Format:         &audio.Format{SampleRate: sampleRate, NumChannels: 1},
		SourceBitDepth: bitDepth,
	}

	for i, sample := range wave {
		buffer.Data[i] = int(sample)
	}

	encoder := wav.NewEncoder(file, sampleRate, bitDepth, 1, 1)
	defer encoder.Close()

	if err := encoder.Write(buffer); err != nil {
		return fmt.Errorf("error writing to WAV: %v", err)
	}

	return nil
}

func PlaySineWaveSDL(deviceID sdl.AudioDeviceID, wave []int16) error {
	byteData := make([]byte, len(wave)*2)
	for i := 0; i < len(wave); i++ {
		byteData[2*i] = byte(wave[i] & 0xff)
		byteData[2*i+1] = byte((wave[i] >> 8) & 0xff)
	}

	sdl.ClearQueuedAudio(deviceID)

	if err := sdl.QueueAudio(deviceID, byteData); err != nil {
		return fmt.Errorf("Failed to queue audio: %v", err)
	}

	sdl.PauseAudioDevice(deviceID, false)

	// Wait until the audio is done playing
	time.Sleep(duration)

	return nil
}

func main() {
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		log.Fatalf("Failed to initialize SDL2: %v", err)
	}
	defer sdl.Quit()

	// Open the audio device once
	spec := sdl.AudioSpec{
		Freq:     int32(sampleRate44), // Start with 44.1 kHz
		Format:   sdl.AUDIO_S16SYS,    // 16-bit signed
		Channels: 1,                   // Mono
		Samples:  4096,                // Buffer size
	}

	deviceID, err := sdl.OpenAudioDevice("", false, &spec, nil, 0)
	if err != nil {
		log.Fatalf("Failed to open SDL audio device: %v", err)
	}
	defer sdl.CloseAudioDevice(deviceID)

	// Generate 220 Hz sine wave at 44.1 kHz
	fmt.Printf("Playing %.0f Hz sine wave at 44.1 kHz\n", frequency1)
	sineWave44 := GenerateSineWave(frequency1, sampleRate44, duration)
	if err := PlaySineWaveSDL(deviceID, sineWave44); err != nil {
		log.Fatalf("Failed to play audio: %v", err)
	}

	// Change sample rate for the second sine wave
	spec.Freq = int32(sampleRate48)

	// Clear previously queued audio
	sdl.ClearQueuedAudio(deviceID)

	fmt.Printf("Playing %.0f Hz sine wave at 48 kHz\n", frequency2)
	sineWave48 := GenerateSineWave(frequency2, sampleRate48, duration)
	if err := PlaySineWaveSDL(deviceID, sineWave48); err != nil {
		log.Fatalf("Failed to play audio: %v", err)
	}

	filename44 := fmt.Sprintf("%dHz_%s_44k.wav", int(frequency1), waveform)
	fmt.Println("Writing sine wave at 44.1 kHz to", filename44)
	if err := WriteWaveToFile(filename44, sineWave44, sampleRate44); err != nil {
		log.Fatalf("Failed to write %s: %v", filename44, err)
	}

	filename48 := fmt.Sprintf("%dHz_%s_48k.wav", int(frequency2), waveform)
	fmt.Println("Writing sine wave at 48 kHz to", filename48)
	if err := WriteWaveToFile(filename48, sineWave48, sampleRate48); err != nil {
		log.Fatalf("Failed to write %s: %v", filename48, err)
	}
}
