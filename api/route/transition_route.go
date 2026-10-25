package route

import (
	"github.com/DIEALOG/DIEALOG-backend/api/handler"
	"github.com/DIEALOG/DIEALOG-backend/api/middleware"
	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/gofiber/fiber/v3"
)

func NewTransitionRouter(
	app fiber.Router,
	policyService domain.UserTransitionPolicyUseCase,
	confirmerService domain.UserTransitionConfirmerUseCase,
	aliveCheckService domain.UserAliveCheckUseCase,
	caseService domain.UserTransitionCaseUseCase,
	confirmationService domain.UserTransitionConfirmationUseCase,
	accountHandlingService domain.AccountHandlingUseCase,
) {
	// owner JWT routes
	protected := app.Group("protected")
	protected.Use(middleware.JwtMiddleware)

	protected.Get("/policy", handler.FetchTransitionPolicy(policyService))
	protected.Patch("/policy", handler.UpdateTransitionPolicy(policyService))

	protected.Get("/confirmers", handler.FetchTransitionConfirmers(confirmerService))
	protected.Post("/confirmers", handler.InviteTransitionConfirmer(confirmerService))
	protected.Post("/confirmers/:id/resend", handler.ResendTransitionConfirmerInvitation(confirmerService))
	protected.Delete("/confirmers/:id", handler.RevokeTransitionConfirmer(confirmerService))

	protected.Get("/alive-checks", handler.FetchAliveChecks(aliveCheckService))
	protected.Post("/alive-checks", handler.ScheduleAliveCheck(aliveCheckService))
	protected.Post("/alive-checks/:id/respond", handler.RespondAliveCheck(aliveCheckService))

	protected.Get("/cases", handler.FetchTransitionCases(caseService))
	protected.Get("/cases/:id", handler.FetchTransitionCase(caseService))
	protected.Post("/cases/:id/cancel", handler.CancelTransitionCase(caseService))
	protected.Get("/cases/:id/confirmations", handler.FetchTransitionConfirmations(confirmationService))
	protected.Get("/cases/:id/tasks", handler.FetchAccountHandlingTasks(accountHandlingService))
	protected.Post("/tasks/:id/request", handler.RequestAccountHandlingTask(accountHandlingService))

	// confirmer token routes (separate from owner JWT auth)
	confirmer := app.Group("confirmer")
	confirmer.Use(middleware.ConfirmerMiddleware(confirmerService))

	confirmer.Get("/cases", handler.FetchConfirmerCases(caseService))
	confirmer.Post("/cases", handler.StartTransitionCase(caseService))
	confirmer.Get("/cases/:id", handler.FetchConfirmerCase(caseService))
	confirmer.Post("/cases/:id/confirmations", handler.SubmitTransitionConfirmation(confirmationService))

	// public invitation token routes. The token is sent in the JSON body so it
	// does not appear in access logs.
	invitations := app.Group("invitations")
	invitations.Post("/accept", handler.AcceptTransitionInvitation(confirmerService))
	invitations.Post("/decline", handler.DeclineTransitionInvitation(confirmerService))

	// public stakeholder token route. The token is sent in the JSON body so it
	// does not appear in access logs.
	stakeholderTasks := app.Group("stakeholder-tasks")
	stakeholderTasks.Post("/result", handler.SubmitAccountHandlingResult(accountHandlingService))
}
