package admin

import (
	"context"
	"github.com/CMSgov/dpc/api/client"
	"github.com/CMSgov/dpc/api/conf"
	middleware2 "github.com/CMSgov/dpc/api/middleware"
	"github.com/CMSgov/dpc/api/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"net/http"

	v2 "github.com/CMSgov/dpc/api/v2"
	"github.com/go-chi/chi/middleware"
)

func buildAdminRoutes(c controllers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware2.Logging())
	r.Use(middleware2.RequestIPCtx)
	r.With(middleware2.Sanitize).Route("/api/v2", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/fhir+json; charset=UTF-8"))
		r.Get("/_health", func(w http.ResponseWriter, r *http.Request) {
			m := make(map[string]string)
			m["api"] = "ok"
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, m)
		})

		//ORGANIZATION Routes
		r.Route("/Organization", func(r chi.Router) {
			r.Route("/{organizationID}", func(r chi.Router) {
				r.Use(middleware2.AdminOrganizationCtx)
				r.With(middleware2.FHIRModel).Get("/", c.Org.Read)
				r.Delete("/", c.Org.Delete)
				r.With(middleware2.FHIRFilter, middleware2.FHIRModel).Put("/", c.Org.Update)
			})
			r.With(middleware2.FHIRFilter, middleware2.FHIRModel).Post("/", c.Org.Create)
		})

		//IMPLEMENTER Routes
		r.Route("/Implementer", func(r chi.Router) {
			r.Post("/", c.Impl.Create)
			r.Route("/{implementerID}/org", func(r chi.Router) {
				r.Use(middleware2.ImplementerCtx)
				r.Get("/", c.ImplOrg.Read)
				r.Post("/", c.ImplOrg.Create)
			})
		})
		//IMPLEMENTER ORG
		r.Route("/Implementer/{implementerID}/Org/{organizationID}/system", func(r chi.Router) {
			r.With(middleware2.ImplementerCtx).With(middleware2.AdminOrganizationCtx).Post("/", c.Ssas.CreateSystem)
			r.With(middleware2.ImplementerCtx, middleware2.AdminOrganizationCtx).Get("/", c.Ssas.GetSystem)
		})

		r.Route("/Implementer/{implementerID}/Org/{organizationID}", func(r chi.Router) {
			r.With(middleware2.ImplementerCtx, middleware2.AdminOrganizationCtx).Post("/token", c.Ssas.CreateToken)
			r.With(middleware2.ImplementerCtx, middleware2.AdminOrganizationCtx, middleware2.TokenCtx).Delete("/token/{tokenID}", c.Ssas.DeleteToken)
			r.With(middleware2.ImplementerCtx, middleware2.AdminOrganizationCtx).Post("/key", c.Ssas.AddKey)
			r.With(middleware2.ImplementerCtx, middleware2.AdminOrganizationCtx, middleware2.PublicKeyCtx).Delete("/key/{keyID}", c.Ssas.DeleteKey)
		})

	})
	return r
}

// NewAdminServer configures clients, builds ADMIN routes, and creates a server.
func NewAdminServer(ctx context.Context) *service.Server {
	attrClient := client.NewAttributionClient(ctx, client.AttributionConfig{
		URL:     conf.GetAsString("attribution-client.url"),
		Retries: conf.GetAsInt("attribution-client.retries", 3),
		CACert:  conf.GetAsString("ATTR_CA_CERT"),
		Cert:    conf.GetAsString("ATTR_CERT"),
		CertKey: conf.GetAsString("ATTR_CERT_KEY"),
	})

	ssasClient := client.NewSsasHTTPClient(ctx, client.SsasHTTPClientConfig{
		PublicURL:    conf.GetAsString("ssas-client.public-url"),
		AdminURL:     conf.GetAsString("ssas-client.admin-url"),
		Retries:      conf.GetAsInt("ssas-client.attrRetries", 3),
		ClientID:     conf.GetAsString("ssas-client.client-id"),
		ClientSecret: conf.GetAsString("ssas-client.client-secret"),
        CACert: conf.GetAsString("ssas-client.ca-cert"),
        Cert: conf.GetAsString("ssas-client.cert"),
        CertKey: conf.GetAsString("ssas-client.cert-key"),
	})

	port := conf.GetAsInt("ADMIN_PORT", 3011)

	controllers := controllers{
		Org:     v2.NewOrganizationController(attrClient),
		Impl:    v2.NewImplementerController(attrClient, ssasClient),
		ImplOrg: v2.NewImplementerOrgController(attrClient),
		Ssas:    v2.NewSSASController(ssasClient, attrClient),
	}

	r := buildAdminRoutes(controllers)
	authType := conf.GetAsString("AUTH_TYPE", "TLS")
	return service.NewServer("DPC-API Admin Server", port, authType, r)
}

type controllers struct {
	Org     v2.Controller
	Health  v2.Controller
	Impl    v2.Controller
	ImplOrg v2.Controller
	Ssas    v2.AuthController
}
