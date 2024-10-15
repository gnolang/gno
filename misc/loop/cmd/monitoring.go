package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// Monitoring to export
// - Number of lines in backups.jsonl
// - Uptime of last version

type monitoringService struct {
	portalLoop *snapshotter
}

type metrics struct {
	backupTXS prometheus.Gauge
}

func (s *monitoringService) getBackupsFileMetrics(m *metrics) error {
	f, err := os.Open(s.portalLoop.backupFile)
	if err != nil {
		return fmt.Errorf("monitoring: %w, file: %s", err, s.portalLoop.backupFile)
	}

	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)

	backedUpTxs := 0
	for fileScanner.Scan() {
		backedUpTxs += 1

		// TODO(albttx): Increment a counter if tx is a new realm
		// content := fileScanner.Text()
		// fmt.Println(content)
	}

	m.backupTXS.Set(float64(backedUpTxs))

	return nil
}

func (s *monitoringService) recordMetrics() {
	m := metrics{}

	m.backupTXS = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "portal_loop_backup_txs",
		Help: "The number of line backed up",
		ConstLabels: prometheus.Labels{
			"portal_loop_container_name": s.portalLoop.containerName,
		},
	})

	// Start infinite loop of metrics getter
	for {
		if err := s.getBackupsFileMetrics(&m); err != nil {
			logrus.WithError(err).Error()
		}

		time.Sleep(time.Second * 5)
	}
}
