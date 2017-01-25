package transaction

import (
	"testing"

	"golang.org/x/net/context"

	transactionid "github.com/giantswarm/microkit/transaction/context/id"
)

func Test_Executer_NoTransactionIDGiven(t *testing.T) {
	config := DefaultExecuterConfig()
	newExecuter, err := NewExecuter(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	var replayExecuted int
	var trialExecuted int

	replay := func(context context.Context) error {
		replayExecuted++
		return nil
	}
	trial := func(context context.Context) error {
		trialExecuted++
		return nil
	}

	var ctx context.Context
	var executeConfig ExecuteConfig
	{
		ctx = context.Background()

		executeConfig = newExecuter.ExecuteConfig()
		executeConfig.Replay = replay
		executeConfig.Trial = trial
		executeConfig.TrialID = "test-trial-ID"
	}

	// The first execution of the transaction causes the trial to be executed
	// once. The replay function must not be executed at all.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}

	// There is no transaction ID provided, so the trial is executed again and the
	// replay function is still untouched.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 2 {
			t.Fatal("expected", 2, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}

	// There is no transaction ID provided, so the trial is executed again and the
	// replay function is still untouched.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 3 {
			t.Fatal("expected", 3, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}
}

func Test_Executer_TransactionIDGiven(t *testing.T) {
	config := DefaultExecuterConfig()
	newExecuter, err := NewExecuter(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	var replayExecuted int
	var trialExecuted int

	replay := func(context context.Context) error {
		replayExecuted++
		return nil
	}
	trial := func(context context.Context) error {
		trialExecuted++
		return nil
	}

	var ctx context.Context
	var executeConfig ExecuteConfig
	{
		ctx = context.Background()
		ctx = transactionid.NewContext(ctx, "test-transaction-id")

		executeConfig = newExecuter.ExecuteConfig()
		executeConfig.Replay = replay
		executeConfig.Trial = trial
		executeConfig.TrialID = "test-trial-ID"
	}

	// The first execution of the transaction causes the trial to be executed
	// once. The replay function must not be executed at all.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}

	// There is a transaction ID provided, so the trial is not executed again and
	// the replay function is executed the first time.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 1 {
			t.Fatal("expected", 1, "got", replayExecuted)
		}
	}

	// There is a transaction ID provided, so the trial is still not executed
	// again and the replay function is executed the second time.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 2 {
			t.Fatal("expected", 2, "got", replayExecuted)
		}
	}
}
