package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	httpSwagger "github.com/swaggo/http-swagger/v2" // http-swagger middleware
	"github.com/sysadminsmedia/homebox/backend/app/api/handlers/debughandlers"
	v1 "github.com/sysadminsmedia/homebox/backend/app/api/handlers/v1"
	"github.com/sysadminsmedia/homebox/backend/app/api/providers"
	docs "github.com/sysadminsmedia/homebox/backend/app/api/static/docs"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authroles"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

const prefix = "/api"

var (
	ErrDir = errors.New("path is dir")

	//go:embed all:static/public/*
	public embed.FS
)

func (a *app) debugRouter() *http.ServeMux {
	dbg := http.NewServeMux()
	debughandlers.New(dbg)

	return dbg
}

// swaggerDocHandler serves the OpenAPI document, rewriting its `host` so the
// Swagger UI's "Base URL" reflects the instance actually being used rather than
// a value baked in at build time.
//
// That host is the target of every "Try it out" request the Swagger UI issues,
// Authorization header included, and this route is unauthenticated. Honoring a
// raw X-Forwarded-Host therefore let any unauthenticated client mint a spec
// pointing wherever they liked; behind a cache that does not key on that header,
// the poisoned document would then be served to operators. Defer to
// SecureBaseURL, which prefers the operator's configured hostname and accepts
// X-Forwarded-Host only under TrustProxy, and only after validation. The
// fallback is r.Host, which caches already key on.
func (a *app) swaggerDocHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if trusted := v1.SecureBaseURL(r, &a.conf.Options); trusted != "" {
		if u, err := url.Parse(trusted); err == nil && u.Host != "" {
			host = u.Host
		}
	}

	spec := *docs.SwaggerInfo
	spec.Host = host
	doc := spec.ReadDoc()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(doc))
}

// registerRoutes registers all the routes for the API
func (a *app) mountRoutes(r *chi.Mux, chain *errchain.ErrChain, repos *repo.AllRepos) {
	registerMimes()

	r.Get("/swagger/doc.json", a.swaggerDocHandler)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// =========================================================================
	// API Version 1

	v1Ctrl := v1.NewControllerV1(
		a.services,
		a.repos,
		a.bus,
		a.conf,
		v1.WithMaxUploadSize(a.conf.Web.MaxUploadSize),
		v1.WithMaxImportSize(a.conf.Web.MaxImportSize),
		v1.WithMaxParseMemory(a.conf.Web.MaxParseMemory),
		v1.WithRegistration(a.conf.Options.AllowRegistration),
		v1.WithDemoStatus(a.conf.Demo), // Disable Password Change in Demo Mode
		v1.WithURL(fmt.Sprintf("%s:%s", a.conf.Web.Host, a.conf.Web.Port)),
	)

	r.Route(prefix+"/v1", func(r chi.Router) {
		r.Get("/status", chain.ToHandlerFunc(v1Ctrl.HandleBase(func() bool { return true }, v1.Build{
			Version:   version,
			Commit:    commit,
			BuildTime: buildTime,
		})))

		r.Get("/currencies", chain.ToHandlerFunc(v1Ctrl.HandleCurrency()))

		providers := []v1.AuthProvider{
			providers.NewLocalProvider(a.services.User),
		}

		r.Post("/users/register", chain.ToHandlerFunc(v1Ctrl.HandleUserRegistration(), a.mwRegisterRateLimit))
		r.Post("/users/login", chain.ToHandlerFunc(v1Ctrl.HandleAuthLogin(providers...), a.mwAuthRateLimit))
		r.Post("/users/forgot-password", chain.ToHandlerFunc(v1Ctrl.HandleForgotPassword(), a.mwAuthRateLimit))
		r.Post("/users/reset-password", chain.ToHandlerFunc(v1Ctrl.HandleResetPassword(), a.mwAuthRateLimit))

		if a.conf.OIDC.Enabled {
			r.Get("/users/login/oidc", chain.ToHandlerFunc(v1Ctrl.HandleOIDCLogin(), a.mwAuthRateLimit))
			r.Get("/users/login/oidc/callback", chain.ToHandlerFunc(v1Ctrl.HandleOIDCCallback(), a.mwAuthRateLimit))
		}

		userMW := []errchain.Middleware{
			a.mwAuthToken,
			a.mwTenant,
			a.mwRoles(RoleModeOr, authroles.RoleUser.String(), authroles.RoleAdmin.String()),
		}

		adminMW := []errchain.Middleware{
			a.mwAuthToken,
			a.mwTenant,
			a.mwRoles(RoleModeOr, authroles.RoleAdmin.String()),
		}

		// ownerMW additionally requires role=owner on the tenant collection.
		// Use it for anything that administers the collection itself rather
		// than its contents — an invited member must not be able to rename,
		// re-share, or delete a collection they were merely given access to.
		ownerMW := []errchain.Middleware{
			a.mwAuthToken,
			a.mwTenant,
			a.mwRoles(RoleModeOr, authroles.RoleAdmin.String()),
			a.mwGroupOwner,
		}

		r.Get("/ws/events", chain.ToHandlerFunc(v1Ctrl.HandleCacheWS(), append([]errchain.Middleware{a.mwAllowQuerySessionToken}, userMW...)...))

		// User management endpoints
		r.Get("/users/self", chain.ToHandlerFunc(v1Ctrl.HandleUserSelf(), userMW...))
		r.Put("/users/self", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfUpdate(), userMW...))
		r.Delete("/users/self", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfDelete(), userMW...))
		r.Get("/users/self/settings", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfSettingsGet(), userMW...))
		r.Put("/users/self/settings", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfSettingsUpdate(), userMW...))
		r.Post("/users/logout", chain.ToHandlerFunc(v1Ctrl.HandleAuthLogout(), userMW...))
		r.Post("/users/logout/all", chain.ToHandlerFunc(v1Ctrl.HandleAuthLogoutAll(), userMW...))
		r.Get("/users/refresh", chain.ToHandlerFunc(v1Ctrl.HandleAuthRefresh(), userMW...))
		r.Put("/users/self/change-password", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfChangePassword(), userMW...))

		// User API keys (static tokens that authenticate as the owning user)
		r.Get("/users/self/api-keys", chain.ToHandlerFunc(v1Ctrl.HandleUserAPIKeysList(), userMW...))
		r.Post("/users/self/api-keys", chain.ToHandlerFunc(v1Ctrl.HandleUserAPIKeyCreate(), userMW...))
		r.Delete("/users/self/api-keys/{id}", chain.ToHandlerFunc(v1Ctrl.HandleUserAPIKeyDelete(), userMW...))

		// Account roles and supervisors
		r.Get("/account-members", chain.ToHandlerFunc(v1Ctrl.HandleAccountMembersGetAll(), adminMW...))
		r.Get("/account-supervisors", chain.ToHandlerFunc(v1Ctrl.HandleAccountSupervisorsGetAll(), adminMW...))
		r.Post("/account-members", chain.ToHandlerFunc(v1Ctrl.HandleAccountCreate(), adminMW...))
		r.Put("/account-members/{id}/role", chain.ToHandlerFunc(v1Ctrl.HandleAccountRoleUpdate(), adminMW...))
		r.Put("/account-members/{id}/supervisors", chain.ToHandlerFunc(v1Ctrl.HandleAccountSupervisorsUpdate(), adminMW...))

		// Group management endpoints
		r.Get("/groups/all", chain.ToHandlerFunc(v1Ctrl.HandleGroupsGetAll(), userMW...))
		r.Post("/groups", chain.ToHandlerFunc(v1Ctrl.HandleGroupCreate(), adminMW...))
		r.Get("/groups", chain.ToHandlerFunc(v1Ctrl.HandleGroupGet(), userMW...))
		r.Put("/groups", chain.ToHandlerFunc(v1Ctrl.HandleGroupUpdate(), ownerMW...))
		r.Delete("/groups", chain.ToHandlerFunc(v1Ctrl.HandleGroupDelete(), ownerMW...))

		r.Get("/groups/members", chain.ToHandlerFunc(v1Ctrl.HandleGroupMembersGetAll(), userMW...))
		r.Delete("/groups/members/{user_id}", chain.ToHandlerFunc(v1Ctrl.HandleGroupMemberRemove(), ownerMW...))

		r.Get("/groups/invitations", chain.ToHandlerFunc(v1Ctrl.HandleGroupInvitationsGetAll(), userMW...))
		r.Post("/groups/invitations", chain.ToHandlerFunc(v1Ctrl.HandleGroupInvitationsCreate(), ownerMW...))
		r.Delete("/groups/invitations/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGroupInvitationsDelete(), ownerMW...))
		r.Post("/groups/invitations/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGroupInvitationsAccept(), userMW...))

		// Collection export/import (group-scoped)
		r.Post("/group/exports", chain.ToHandlerFunc(v1Ctrl.HandleExportsCreate(), adminMW...))
		r.Get("/group/exports", chain.ToHandlerFunc(v1Ctrl.HandleExportsList(), userMW...))
		r.Get("/group/exports/{id}", chain.ToHandlerFunc(v1Ctrl.HandleExportGet(), userMW...))
		r.Get("/group/exports/{id}/download", chain.ToHandlerFunc(v1Ctrl.HandleExportDownload(), userMW...))
		r.Delete("/group/exports/{id}", chain.ToHandlerFunc(v1Ctrl.HandleExportDelete(), userMW...))
		r.Post("/group/import", chain.ToHandlerFunc(v1Ctrl.HandleCollectionImport(), adminMW...))

		r.Get("/groups/statistics", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatistics(), userMW...))
		r.Get("/groups/statistics/purchase-price", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsPriceOverTime(), userMW...))
		r.Get("/groups/statistics/locations", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsLocations(), userMW...))
		r.Get("/groups/statistics/tags", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsTags(), userMW...))

		// Action endpoints
		r.Post("/actions/ensure-asset-ids", chain.ToHandlerFunc(v1Ctrl.HandleEnsureAssetID(), adminMW...))
		r.Post("/actions/zero-item-time-fields", chain.ToHandlerFunc(v1Ctrl.HandleItemDateZeroOut(), adminMW...))
		r.Post("/actions/ensure-import-refs", chain.ToHandlerFunc(v1Ctrl.HandleEnsureImportRefs(), adminMW...))
		r.Post("/actions/set-primary-photos", chain.ToHandlerFunc(v1Ctrl.HandleSetPrimaryPhotos(), adminMW...))
		r.Post("/actions/create-missing-thumbnails", chain.ToHandlerFunc(v1Ctrl.HandleCreateMissingThumbnails(), adminMW...))
		r.Post("/actions/wipe-inventory", chain.ToHandlerFunc(v1Ctrl.HandleWipeInventory(), adminMW...))

		// Tags endpoints
		r.Get("/tags", chain.ToHandlerFunc(v1Ctrl.HandleTagsGetAll(), userMW...))
		r.Post("/tags", chain.ToHandlerFunc(v1Ctrl.HandleTagsCreate(), adminMW...))
		r.Get("/tags/{id}", chain.ToHandlerFunc(v1Ctrl.HandleTagGet(), userMW...))
		r.Put("/tags/{id}", chain.ToHandlerFunc(v1Ctrl.HandleTagUpdate(), adminMW...))
		r.Delete("/tags/{id}", chain.ToHandlerFunc(v1Ctrl.HandleTagDelete(), adminMW...))

		// Entity Type endpoints
		r.Get("/entity-types", chain.ToHandlerFunc(v1Ctrl.HandleEntityTypeGetAll(), userMW...))
		r.Post("/entity-types", chain.ToHandlerFunc(v1Ctrl.HandleEntityTypeCreate(), adminMW...))
		r.Put("/entity-types/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityTypeUpdate(), adminMW...))
		r.Delete("/entity-types/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityTypeDelete(), adminMW...))

		// Entity endpoints (primary)
		r.Get("/entities", chain.ToHandlerFunc(v1Ctrl.HandleEntitiesGetAll(), userMW...))
		r.Post("/entities", chain.ToHandlerFunc(v1Ctrl.HandleEntitiesCreate(), adminMW...))
		r.Post("/entities/import", chain.ToHandlerFunc(v1Ctrl.HandleEntitiesImport(), adminMW...))
		r.Get("/entities/export", chain.ToHandlerFunc(v1Ctrl.HandleEntitiesExport(), userMW...))
		r.Get("/entities/fields", chain.ToHandlerFunc(v1Ctrl.HandleGetAllCustomFieldNames(), userMW...))
		r.Get("/entities/fields/values", chain.ToHandlerFunc(v1Ctrl.HandleGetAllCustomFieldValues(), userMW...))
		r.Get("/entities/tree", chain.ToHandlerFunc(v1Ctrl.HandleLocationTreeQuery(), userMW...))

		r.Get("/entities/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityGet(), userMW...))
		r.Get("/entities/{id}/path", chain.ToHandlerFunc(v1Ctrl.HandleEntityFullPath(), userMW...))
		r.Put("/entities/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityUpdate(), adminMW...))
		r.Patch("/entities/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityPatch(), adminMW...))
		r.Delete("/entities/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityDelete(), adminMW...))
		r.Post("/entities/{id}/duplicate", chain.ToHandlerFunc(v1Ctrl.HandleEntityDuplicate(), adminMW...))

		// Inventory recipients, assignments, and append-only movement log
		r.Get("/inventory/recipients", chain.ToHandlerFunc(v1Ctrl.HandleInventoryRecipientsGetAll(), userMW...))
		r.Post("/inventory/recipients", chain.ToHandlerFunc(v1Ctrl.HandleInventoryRecipientCreate(), adminMW...))
		r.Get("/inventory/recipients/{id}", chain.ToHandlerFunc(v1Ctrl.HandleInventoryRecipientGet(), userMW...))
		r.Put("/inventory/recipients/{id}", chain.ToHandlerFunc(v1Ctrl.HandleInventoryRecipientUpdate(), adminMW...))
		r.Post("/inventory/recipients/{id}/deactivate", chain.ToHandlerFunc(v1Ctrl.HandleInventoryRecipientDeactivate(), adminMW...))
		r.Get("/inventory/recipients/{id}/transactions", chain.ToHandlerFunc(v1Ctrl.HandleInventoryRecipientTransactions(), userMW...))
		r.Post("/inventory/issue", chain.ToHandlerFunc(v1Ctrl.HandleInventoryIssue(), adminMW...))
		r.Post("/inventory/stock", chain.ToHandlerFunc(v1Ctrl.HandleInventoryStockAdd(), adminMW...))
		r.Post("/inventory/write-off", chain.ToHandlerFunc(v1Ctrl.HandleInventoryStockWriteOff(), adminMW...))
		r.Get("/inventory/assignments", chain.ToHandlerFunc(v1Ctrl.HandleInventoryAssignmentsGetAll(), userMW...))
		r.Post("/inventory/assignments/{id}/return", chain.ToHandlerFunc(v1Ctrl.HandleInventoryReturn(), adminMW...))
		r.Get("/inventory/items/{id}/transactions", chain.ToHandlerFunc(v1Ctrl.HandleInventoryItemTransactions(), userMW...))
		r.Get("/inventory/transactions/export", chain.ToHandlerFunc(v1Ctrl.HandleInventoryTransactionsCSV(), userMW...))
		r.Get("/inventory/transactions", chain.ToHandlerFunc(v1Ctrl.HandleInventoryTransactionsGetAll(), userMW...))

		// Entity attachment endpoints
		r.Post("/entities/{id}/attachments", chain.ToHandlerFunc(v1Ctrl.HandleEntityAttachmentCreate(), adminMW...))
		r.Post("/entities/{id}/attachments/external", chain.ToHandlerFunc(v1Ctrl.HandleEntityAttachmentExternalCreate(), adminMW...))
		r.Put("/entities/{id}/attachments/{attachment_id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityAttachmentUpdate(), adminMW...))
		r.Delete("/entities/{id}/attachments/{attachment_id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityAttachmentDelete(), adminMW...))

		// Entity maintenance endpoints
		r.Get("/entities/{id}/maintenance", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceLogGet(), userMW...))
		r.Post("/entities/{id}/maintenance", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryCreate(), adminMW...))

		r.Get("/assets/{id}", chain.ToHandlerFunc(v1Ctrl.HandleAssetGet(), userMW...))

		// Entity Templates
		r.Get("/templates", chain.ToHandlerFunc(v1Ctrl.HandleEntityTemplatesGetAll(), userMW...))
		r.Post("/templates", chain.ToHandlerFunc(v1Ctrl.HandleEntityTemplatesCreate(), adminMW...))
		r.Get("/templates/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityTemplatesGet(), userMW...))
		r.Put("/templates/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityTemplatesUpdate(), adminMW...))
		r.Delete("/templates/{id}", chain.ToHandlerFunc(v1Ctrl.HandleEntityTemplatesDelete(), adminMW...))
		r.Post("/templates/{id}/create-item", chain.ToHandlerFunc(v1Ctrl.HandleEntityTemplatesCreateItem(), adminMW...))

		// Label Templates
		r.Get("/label-templates", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesGetAll(), userMW...))
		r.Post("/label-templates", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesCreate(), adminMW...))
		r.Get("/label-templates/presets", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesPresets(), userMW...))
		r.Get("/label-templates/barcode-formats", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesBarcodeFormats(), userMW...))
		r.Get("/label-templates/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesGet(), userMW...))
		r.Put("/label-templates/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesUpdate(), adminMW...))
		r.Delete("/label-templates/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesDelete(), adminMW...))
		r.Post("/label-templates/{id}/duplicate", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesDuplicate(), adminMW...))
		r.Get("/label-templates/{id}/preview", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesPreview(), userMW...))
		r.Post("/label-templates/{id}/render", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesRender(), userMW...))
		r.Post("/label-templates/{id}/print", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesPrint(), userMW...))
		r.Post("/label-templates/{id}/render-assets", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesRenderAssets(), userMW...))
		r.Post("/label-templates/{id}/render-locations", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesRenderLocations(), userMW...))
		r.Post("/label-templates/{id}/print-locations", chain.ToHandlerFunc(v1Ctrl.HandleLabelTemplatesPrintLocations(), userMW...))

		// Maintenance
		r.Get("/maintenance", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceGetAll(), userMW...))
		r.Put("/maintenance/{id}", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryUpdate(), adminMW...))
		r.Delete("/maintenance/{id}", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryDelete(), adminMW...))

		// Notifiers
		r.Get("/notifiers", chain.ToHandlerFunc(v1Ctrl.HandleGetUserNotifiers(), userMW...))
		r.Post("/notifiers", chain.ToHandlerFunc(v1Ctrl.HandleCreateNotifier(), adminMW...))
		r.Put("/notifiers/{id}", chain.ToHandlerFunc(v1Ctrl.HandleUpdateNotifier(), adminMW...))
		r.Delete("/notifiers/{id}", chain.ToHandlerFunc(v1Ctrl.HandleDeleteNotifier(), adminMW...))
		r.Post("/notifiers/test", chain.ToHandlerFunc(v1Ctrl.HandlerNotifierTest(), append(adminMW, a.notifierTestLimiter.middleware)...))

		// Asset-Like endpoints
		assetMW := []errchain.Middleware{
			a.mwAuthToken,
			a.mwTenant,
			a.mwRoles(RoleModeOr, authroles.RoleUser.String(), authroles.RoleAttachments.String()),
		}

		r.Get("/products/search-from-barcode", chain.ToHandlerFunc(v1Ctrl.HandleProductSearchFromBarcode(a.conf.Barcode), userMW...))

		r.Get("/qrcode", chain.ToHandlerFunc(v1Ctrl.HandleGenerateQRCode(), assetMW...))
		r.Get(
			"/entities/{id}/attachments/{attachment_id}",
			chain.ToHandlerFunc(v1Ctrl.HandleEntityAttachmentGet(), assetMW...),
		)

		// Labelmaker
		r.Get("/labelmaker/entity/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGetItemLabel(), userMW...))
		r.Get("/labelmaker/location/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGetLocationLabel(), userMW...))
		r.Get("/labelmaker/item/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGetItemLabel(), userMW...))
		r.Get("/labelmaker/asset/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGetAssetLabel(), userMW...))

		// Reporting Services
		r.Get("/reporting/bill-of-materials", chain.ToHandlerFunc(v1Ctrl.HandleBillOfMaterialsExport(), userMW...))

		// OpenTelemetry proxy endpoint for frontend telemetry (requires auth)
		if a.otel != nil && a.otel.IsEnabled() && a.conf.Otel.ProxyEnabled {
			r.Post("/telemetry", chain.ToHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				a.otel.ProxyHandler().ServeHTTP(w, r)
				return nil
			}, userMW...))
		}

		r.NotFound(http.NotFound)
	})

	r.NotFound(chain.ToHandlerFunc(notFoundHandler()))
}

func registerMimes() {
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		panic(err)
	}

	err = mime.AddExtensionType(".mjs", "application/javascript")
	if err != nil {
		panic(err)
	}
}

// notFoundHandler perform the main logic around handling the internal SPA embed and ensuring that
// the client side routing is handled correctly.
func notFoundHandler() errchain.HandlerFunc {
	tryRead := func(fs embed.FS, prefix, requestedPath string, w http.ResponseWriter) error {
		f, err := fs.Open(path.Join(prefix, requestedPath))
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		stat, _ := f.Stat()
		if stat.IsDir() {
			return ErrDir
		}

		contentType := mime.TypeByExtension(filepath.Ext(requestedPath))
		w.Header().Set("Content-Type", contentType)
		_, err = io.Copy(w, f)
		return err
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		err := tryRead(public, "static/public", r.URL.Path, w)
		if err != nil {
			// Fallback to the index.html file.
			// should succeed in all cases.
			err = tryRead(public, "static/public", "index.html", w)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
