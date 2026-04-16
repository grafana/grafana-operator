package main

import (
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func TestApplyLeaderElectionTiming_AllZero(t *testing.T) {
	operatorConfig.LeaderElectionLeaseDuration = 0
	operatorConfig.LeaderElectionRenewDeadline = 0
	operatorConfig.LeaderElectionRetryPeriod = 0

	opts := ctrl.Options{}
	if err := applyLeaderElectionTiming(&opts); err != nil {
		t.Fatalf("expected no error for all-zero (unset) values, got: %v", err)
	}
	if opts.LeaseDuration != nil || opts.RenewDeadline != nil || opts.RetryPeriod != nil {
		t.Fatal("expected nil pointers when no overrides are set")
	}
}

func TestApplyLeaderElectionTiming_AllSet(t *testing.T) {
	operatorConfig.LeaderElectionLeaseDuration = 60 * time.Second
	operatorConfig.LeaderElectionRenewDeadline = 40 * time.Second
	operatorConfig.LeaderElectionRetryPeriod = 10 * time.Second

	opts := ctrl.Options{}
	if err := applyLeaderElectionTiming(&opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.LeaseDuration == nil || *opts.LeaseDuration != 60*time.Second {
		t.Fatalf("expected LeaseDuration=60s, got %v", opts.LeaseDuration)
	}
	if opts.RenewDeadline == nil || *opts.RenewDeadline != 40*time.Second {
		t.Fatalf("expected RenewDeadline=40s, got %v", opts.RenewDeadline)
	}
	if opts.RetryPeriod == nil || *opts.RetryPeriod != 10*time.Second {
		t.Fatalf("expected RetryPeriod=10s, got %v", opts.RetryPeriod)
	}
}

func TestApplyLeaderElectionTiming_PartialOverride(t *testing.T) {
	operatorConfig.LeaderElectionLeaseDuration = 30 * time.Second
	operatorConfig.LeaderElectionRenewDeadline = 0
	operatorConfig.LeaderElectionRetryPeriod = 0

	opts := ctrl.Options{}
	if err := applyLeaderElectionTiming(&opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.LeaseDuration == nil || *opts.LeaseDuration != 30*time.Second {
		t.Fatalf("expected LeaseDuration=30s, got %v", opts.LeaseDuration)
	}
	if opts.RenewDeadline != nil {
		t.Fatal("expected RenewDeadline to remain nil for partial override")
	}
	if opts.RetryPeriod != nil {
		t.Fatal("expected RetryPeriod to remain nil for partial override")
	}
}

func TestApplyLeaderElectionTiming_NegativeValue(t *testing.T) {
	operatorConfig.LeaderElectionLeaseDuration = -5 * time.Second
	operatorConfig.LeaderElectionRenewDeadline = 0
	operatorConfig.LeaderElectionRetryPeriod = 0

	opts := ctrl.Options{}
	if err := applyLeaderElectionTiming(&opts); err == nil {
		t.Fatal("expected error for negative lease duration")
	}
}

func TestApplyLeaderElectionTiming_LeaseNotGreaterThanRenew(t *testing.T) {
	operatorConfig.LeaderElectionLeaseDuration = 10 * time.Second
	operatorConfig.LeaderElectionRenewDeadline = 10 * time.Second
	operatorConfig.LeaderElectionRetryPeriod = 2 * time.Second

	opts := ctrl.Options{}
	if err := applyLeaderElectionTiming(&opts); err == nil {
		t.Fatal("expected error when lease duration equals renew deadline")
	}
}

func TestApplyLeaderElectionTiming_RenewNotGreaterThanRetry(t *testing.T) {
	operatorConfig.LeaderElectionLeaseDuration = 30 * time.Second
	operatorConfig.LeaderElectionRenewDeadline = 5 * time.Second
	operatorConfig.LeaderElectionRetryPeriod = 5 * time.Second

	opts := ctrl.Options{}
	if err := applyLeaderElectionTiming(&opts); err == nil {
		t.Fatal("expected error when renew deadline equals retry period")
	}
}
