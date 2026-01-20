package monitor

import (
	"testing"

	"github.com/shirou/gopsutil/v4/disk"
)

func TestMatchMount(t *testing.T) {
	list := []string{"/", "/data*"}
	if !matchMount(list, "/") {
		t.Fatalf("expected root match")
	}
	if !matchMount(list, "/data1") {
		t.Fatalf("expected prefix match")
	}
	if matchMount(list, "/home") {
		t.Fatalf("did not expect /home match")
	}
}

func TestFilterPartitionsIncludeOverridesExclude(t *testing.T) {
	parts := []disk.PartitionStat{
		{Mountpoint: "/", Fstype: "ext4"},
		{Mountpoint: "/tmp", Fstype: "tmpfs"},
	}
	filtered := filterPartitions(parts, FilterConfig{
		MountInclude:  []string{"/"},
		MountExclude:  []string{"/"},
		FstypeExclude: []string{},
	})
	if len(filtered) != 1 || filtered[0].Mountpoint != "/" {
		t.Fatalf("expected include to override exclude, got %#v", filtered)
	}
}

func TestFilterPartitionsFstypeExclude(t *testing.T) {
	parts := []disk.PartitionStat{
		{Mountpoint: "/", Fstype: "ext4"},
		{Mountpoint: "/tmp", Fstype: "tmpfs"},
	}
	filtered := filterPartitions(parts, FilterConfig{
		MountInclude:  nil,
		MountExclude:  nil,
		FstypeExclude: []string{"tmpfs"},
	})
	if len(filtered) != 1 || filtered[0].Mountpoint != "/" {
		t.Fatalf("expected tmpfs excluded, got %#v", filtered)
	}
}
