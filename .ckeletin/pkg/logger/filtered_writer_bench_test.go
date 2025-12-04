package logger

import (
	"testing"

	"github.com/rs/zerolog"
)

// discardWriter is a writer that discards all input (like /dev/null)
type discardWriter struct{}

func (d discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// BenchmarkFilteredWriter_WriteLevel_Pass measures overhead when message passes filter
func BenchmarkFilteredWriter_WriteLevel_Pass(b *testing.B) {
	writer := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.InfoLevel,
	}
	message := []byte("test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = writer.WriteLevel(zerolog.InfoLevel, message)
	}
}

// BenchmarkFilteredWriter_WriteLevel_Filter measures overhead when message is filtered
func BenchmarkFilteredWriter_WriteLevel_Filter(b *testing.B) {
	writer := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.InfoLevel,
	}
	message := []byte("test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = writer.WriteLevel(zerolog.DebugLevel, message)
	}
}

// BenchmarkFilteredWriter_Write measures overhead of plain Write method
func BenchmarkFilteredWriter_Write(b *testing.B) {
	writer := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.InfoLevel,
	}
	message := []byte("test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = writer.Write(message)
	}
}

// BenchmarkDirectWrite_Baseline measures baseline direct write performance
func BenchmarkDirectWrite_Baseline(b *testing.B) {
	writer := discardWriter{}
	message := []byte("test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = writer.Write(message)
	}
}

// BenchmarkDualLogger_InfoToConsole benchmarks dual logger with INFO to console
func BenchmarkDualLogger_InfoToConsole(b *testing.B) {
	consoleWriter := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.InfoLevel,
	}
	fileWriter := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.DebugLevel,
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("test message")
	}
}

// BenchmarkDualLogger_DebugToFileOnly benchmarks dual logger with DEBUG (file only)
func BenchmarkDualLogger_DebugToFileOnly(b *testing.B) {
	consoleWriter := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.InfoLevel,
	}
	fileWriter := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.DebugLevel,
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug().Msg("test message")
	}
}

// BenchmarkSingleLogger_Baseline benchmarks single logger (no filtering)
func BenchmarkSingleLogger_Baseline(b *testing.B) {
	logger := zerolog.New(discardWriter{}).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("test message")
	}
}

// BenchmarkDualLogger_WithFields benchmarks dual logger with structured fields
func BenchmarkDualLogger_WithFields(b *testing.B) {
	consoleWriter := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.InfoLevel,
	}
	fileWriter := FilteredWriter{
		Writer:   discardWriter{},
		MinLevel: zerolog.DebugLevel,
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().
			Str("key1", "value1").
			Int("key2", 42).
			Bool("key3", true).
			Msg("test message")
	}
}

// BenchmarkMultiLevelWriter_NoFilter benchmarks MultiLevelWriter without filtering
func BenchmarkMultiLevelWriter_NoFilter(b *testing.B) {
	multi := zerolog.MultiLevelWriter(discardWriter{}, discardWriter{})
	logger := zerolog.New(multi).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("test message")
	}
}

// BenchmarkMultiLevelWriter_WithFilter benchmarks MultiLevelWriter with filtering
func BenchmarkMultiLevelWriter_WithFilter(b *testing.B) {
	consoleWriter := FilteredWriter{Writer: discardWriter{}, MinLevel: zerolog.InfoLevel}
	fileWriter := FilteredWriter{Writer: discardWriter{}, MinLevel: zerolog.DebugLevel}

	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("test message")
	}
}

// BenchmarkFilteredWriter_DifferentMessageSizes benchmarks different message sizes
func BenchmarkFilteredWriter_DifferentMessageSizes(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		message := make([]byte, size)
		for i := range message {
			message[i] = 'x'
		}

		b.Run(string(rune(size))+"bytes", func(b *testing.B) {
			writer := FilteredWriter{
				Writer:   discardWriter{},
				MinLevel: zerolog.InfoLevel,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = writer.WriteLevel(zerolog.InfoLevel, message)
			}
		})
	}
}
