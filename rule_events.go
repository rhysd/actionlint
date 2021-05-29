package actionlint

import (
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// RuleEvents is a rule to check 'on' field in workflow.
// https://docs.github.com/en/actions/reference/events-that-trigger-workflows
type RuleEvents struct {
	RuleBase
}

// NewRuleEvents creates new RuleEvents instance.
func NewRuleEvents() *RuleEvents {
	return &RuleEvents{
		RuleBase: RuleBase{name: "events"},
	}
}

func (rule *RuleEvents) VisitWorkflowPre(n *Workflow) {
	for _, e := range n.On {
		rule.checkEvent(e)
	}
}

func (rule *RuleEvents) checkEvent(event Event) {
	switch e := event.(type) {
	case *ScheduledEvent:
		for _, c := range e.Cron {
			rule.checkCron(c)
		}
	case *WorkflowDispatchEvent:
		// Nothing to do
	case *RepositoryDispatchEvent:
		// Nothing to do
	case *WebhookEvent:
		rule.checkWebhookEvent(e)
	default:
		panic("unreachable")
	}
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onschedule
func (rule *RuleEvents) checkCron(spec *String) {
	p := cron.NewParser(cron.Month | cron.Hour | cron.Minute | cron.Second | cron.DowOptional)
	sched, err := p.Parse(spec.Value)
	if err != nil {
		rule.errorf(spec.Pos, "invalid CRON format %q in schedule event: %s", spec.Value, err.Error())
		return
	}

	start := time.Unix(0, 0)
	next := sched.Next(start)
	diff := next.Sub(start).Seconds()

	if diff < 60.0 {
		rule.errorf(spec.Pos, "scheduled job runs too frequently. it runs once per %g seconds", diff)
	}
}

// https://docs.github.com/en/actions/reference/events-that-trigger-workflows#webhook-events
func (rule *RuleEvents) checkWebhookEvent(event *WebhookEvent) {
	types := []string{}

	switch event.Hook.Value {
	case "check_run":
		types = []string{"created", "rerequested", "completed"}
	case "check_suite":
		types = []string{"created", "requested", "rerequested"}
	case "create":
	case "delete":
	case "deployment":
	case "deployment_status":
	case "fork":
	case "gollum":
	case "issue_comment":
		types = []string{"created", "edited", "deleted"}
	case "issues":
		types = []string{
			"opened",
			"edited",
			"deleted",
			"transferred",
			"pinned",
			"unpinned",
			"closed",
			"reopened",
			"assigned",
			"unassigned",
			"labeled",
			"unlabeled",
			"locked",
			"unlocked",
			"milestoned",
			"demilestoned",
		}
	case "label":
		types = []string{"created", "edited", "deleted"}
	case "milestone":
		types = []string{"created", "closed", "opened", "edited", "deleted"}
	case "page_build":
	case "project":
		types = []string{"created", "updated", "closed", "reopened", "edited", "deleted"}
	case "project_card":
		types = []string{"created", "moved", "converted", "edited", "deleted"}
	case "project_column":
		types = []string{"created", "updated", "moved", "deleted"}
	case "public":
	case "pull_request":
		types = []string{
			"assigned",
			"unassigned",
			"labeled",
			"unlabeled",
			"opened",
			"edited",
			"closed",
			"reopened",
			"synchronize",
			"ready_for_review",
			"locked",
			"unlocked",
			"review_requested",
			"review_request_removed",
		}
	case "pull_request_review":
		types = []string{"submitted", "edited", "dismissed"}
	case "pull_request_review_comment":
		types = []string{"created", "edited", "deleted"}
	case "pull_request_target":
		types = []string{
			"assigned",
			"unassigned",
			"labeled",
			"unlabeled",
			"opened",
			"edited",
			"closed",
			"reopened",
			"synchronize",
			"ready_for_review",
			"locked",
			"unlocked",
			"review_requested",
			"review_request_removed",
		}
	case "push":
	case "registry_package":
		types = []string{"published", "updated"}
	case "release":
		types = []string{
			"published",
			"unpublished",
			"created",
			"edited",
			"deleted",
			"prereleased",
			"released",
		}
	case "status":
	case "watch":
		types = []string{"started"}
	case "workflow_run":
		types = []string{"completed", "requested"}
		// TODO: Check "workflows" configuration looking at other workflow files
		if len(event.Workflows) == 0 {
			rule.error(event.Pos, "no workflow is configured for \"workflow_run\" event")
		}
	default:
		rule.errorf(event.Pos, "unknown Webhook event %q. see https://docs.github.com/en/actions/reference/events-that-trigger-workflows#webhook-events for list of all Webhook event names", event.Hook.Value)
		return
	}

	rule.checkTypes(event.Hook, event.Types, types)

	if event.Hook.Value != "workflow_run" && len(event.Workflows) != 0 {
		rule.errorf(event.Pos, "\"workflows\" cannot be configured for %q event. it is only for workflow_run event", event.Hook.Value)
	}
}

func (rule *RuleEvents) checkTypes(hook *String, types []*String, expected []string) {
	if len(expected) == 0 && len(types) > 0 {
		rule.errorf(hook.Pos, "\"types\" cannot be specified for %q Webhook event", hook.Value)
		return
	}

	for _, ty := range types {
		valid := false
		for _, e := range expected {
			if ty.Value == e {
				valid = true
				break
			}
		}
		if !valid {
			qs := make([]string, 0, len(expected))
			for _, s := range expected {
				qs = append(qs, strconv.Quote(s))
			}
			rule.errorf(
				ty.Pos,
				"invalid activity type %q for %q Webhook event. available types are %s",
				ty.Value,
				hook.Value,
				strings.Join(qs, ", "),
			)
		}
	}
}
