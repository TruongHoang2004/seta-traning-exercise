package kafka

import (
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// NewWriter tạo kafka.Writer bình thường
func NewWriter(brokers []string, topic string) *kafka.Writer {
	// Ensure topic exists with default configuration
	// Typically, we want at least 3 partitions and replication factor of 2 for production
	err := ensureTopic(brokers, topic, 1, 1) // replicationFactor = 1 cho local
	if err != nil {
		log.Printf("Warning: Failed to ensure topic %s exists: %v", topic, err)
		// Continue anyway as the topic might be created by other means
		// or the error might be due to insufficient permissions
	}

	// time.Sleep(500 * time.Millisecond)

	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

// NewReader tạo kafka.Reader bình thường
func NewReader(brokers []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
		MaxWait:  time.Second,
	})
}

func createTopic(brokers []string, topic string, numPartitions, replicationFactor int) error {
	// Kết nối tới một broker bất kỳ
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	// Lấy controller của cluster
	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	// Kết nối tới controller
	controllerConn, err := kafka.Dial("tcp", controller.Host+":"+strconv.Itoa(controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	// Cấu hình topic
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}

	// Tạo topic
	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		log.Printf("Failed to create topic %s: %v", topic, err)
		return err
	}

	log.Printf("Topic %s created successfully", topic)
	return nil
}

func topicExists(brokers []string, topic string) (bool, error) {
	// Kết nối tới một broker bất kỳ
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// Lấy danh sách topic hiện có
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false, err
	}

	// Duyệt danh sách partition để xem topic có tồn tại
	for _, p := range partitions {
		if p.Topic == topic {
			return true, nil
		}
	}

	return false, nil
}

// ensureTopic kiểm tra topic có tồn tại chưa, nếu chưa thì tạo
func ensureTopic(brokers []string, topic string, numPartitions, replicationFactor int) error {
	exists, err := topicExists(brokers, topic)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("Topic %s already exists", topic)
		return nil
	}

	if err := createTopic(brokers, topic, numPartitions, replicationFactor); err != nil {
		return err
	}

	log.Printf("Topic %s created successfully", topic)
	return nil
}
