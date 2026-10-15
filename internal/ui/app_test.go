package ui

import (
	"github.com/BraveRam/netpulse/internal/killer"
	"github.com/BraveRam/netpulse/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"testing"
	"time"
)

type reviewCollector struct{ conns []*model.Connection }

func (c reviewCollector) GetConnections() ([]*model.Connection, error) { return c.conns, nil }
func (c reviewCollector) GetBandwidth() (*model.BandwidthStats, error) {
	return &model.BandwidthStats{}, nil
}

type recordingKiller struct{ pid int }

func (k *recordingKiller) KillProcess(pid int, _ string, _ bool) killer.Result {
	k.pid = pid
	return killer.Result{Success: true}
}

func TestKillShortcutPinsTargetAcrossRefresh(t *testing.T) {
	m := NewAppModel(Config{})
	defer m.enricher.Close()
	defer m.sniffer.Stop()
	original := &model.Connection{PID: 12345, ProcessName: "original", State: "LISTEN"}
	replacement := &model.Connection{PID: 54321, ProcessName: "replacement", State: "LISTEN"}
	m.connections = []*model.Connection{replacement, original}
	m.applyFilterAndSort()
	m.table.Cursor = 1
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.mode != ViewKill || m.table.Cursor != 1 {
		t.Fatal("k did not open kill dialog on selected row")
	}
	if m.killTarget == nil || m.killTarget.PID != original.PID {
		t.Fatal("wrong initial kill target")
	}
	m.collector = reviewCollector{[]*model.Connection{replacement}}
	m.Update(TickMsg(time.Now()))
	if m.selectedConnection().PID != replacement.PID {
		t.Fatal("refresh did not replace row")
	}
	if m.killTarget.PID != original.PID {
		t.Fatal("refresh changed kill target")
	}
	// Confirm through a fake killer so the test cannot signal a real process.
	recorder := &recordingKiller{}
	m.killer = recorder
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if recorder.pid != original.PID {
		t.Fatalf("confirmation targeted PID %d, want %d", recorder.pid, original.PID)
	}
	if m.mode != ViewNormal {
		t.Fatal("confirmation did not close the dialog")
	}
}

func TestSummaryCountsAreDisjoint(t *testing.T) {
	m := NewAppModel(Config{})
	defer m.enricher.Close()
	defer m.sniffer.Stop()
	m.collector = reviewCollector{[]*model.Connection{
		{Protocol: "tcp", State: "ESTABLISHED"},
		{Protocol: "tcp6", State: "LISTEN"},
		{Protocol: "tcp", State: "TIME_WAIT"},
		{Protocol: "udp", State: "LISTEN"},
		{Protocol: "udp", State: "ESTABLISHED"},
		{Protocol: "udp6", State: "LISTEN"},
		{Protocol: "other", State: "ESTABLISHED"},
	}}
	m.refreshData()
	s := m.stats
	if s.TotalConnections != 7 || s.ActiveTCP != 3 || s.ActiveUDP != 3 || s.ActiveEstablished != 1 || s.ActiveListen != 1 {
		t.Fatalf("overlapping or incorrect counts: %+v", s)
	}
}
