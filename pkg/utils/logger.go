package utils
import (
	"os"
	"sync"
	"github.com/sirupsen/logrus"
)
var (
	logger *logrus.Logger
	once   sync.Once
)
func InitLogger() {
	once.Do(func() {
		logger = logrus.New()
		logger.SetOutput(os.Stdout)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logger.SetLevel(logrus.InfoLevel)
	})
}
func GetLogger() *logrus.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}