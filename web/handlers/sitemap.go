package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type urlsetXML struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapIndexEntry struct {
	Loc string `xml:"loc"`
}

type sitemapIndexXML struct {
	XMLName  xml.Name            `xml:"sitemapindex"`
	Xmlns    string              `xml:"xmlns,attr"`
	Sitemaps []sitemapIndexEntry `xml:"sitemap"`
}

func (s *Server) SitemapIndex(cfg SitemapConfig) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		baseURL := resolveBaseURL(ctx, cfg.BaseURL)

		sm := sitemapIndexXML{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			Sitemaps: []sitemapIndexEntry{
				{Loc: baseURL + "/sitemap-static.xml"},
				{Loc: baseURL + "/sitemap-scrutins.xml"},
				{Loc: baseURL + "/sitemap-deputes.xml"},
				{Loc: baseURL + "/sitemap-groupes.xml"},
			},
		}

		ctx.Response().Header().Set(echo.HeaderContentType, "text/xml; charset=utf-8")
		return ctx.XML(http.StatusOK, sm)
	}
}

func (s *Server) SitemapStatic(cfg SitemapConfig) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		baseURL := resolveBaseURL(ctx, cfg.BaseURL)

		urls := []sitemapURL{
			{Loc: baseURL + "/"},
			{Loc: baseURL + "/scrutins"},
			{Loc: baseURL + "/deputes"},
			{Loc: baseURL + "/groupes"},
		}

		sm := urlsetXML{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		}

		ctx.Response().Header().Set(echo.HeaderContentType, "text/xml; charset=utf-8")
		return ctx.XML(http.StatusOK, sm)
	}
}

func (s *Server) SitemapScrutins(cfg SitemapConfig) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uids, err := s.store.SitemapScrutinUIDs(ctx.Request().Context())
		if err != nil {
			return fmt.Errorf("load scrutin uids: %w", err)
		}

		baseURL := resolveBaseURL(ctx, cfg.BaseURL)
		urls := make([]sitemapURL, 0, len(uids))
		for _, uid := range uids {
			urls = append(urls, sitemapURL{Loc: baseURL + "/scrutins/" + url.PathEscape(uid)})
		}

		sm := urlsetXML{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		}

		ctx.Response().Header().Set(echo.HeaderContentType, "text/xml; charset=utf-8")
		return ctx.XML(http.StatusOK, sm)
	}
}

func (s *Server) SitemapDeputies(cfg SitemapConfig) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uids, err := s.store.SitemapDeputyUIDs(ctx.Request().Context())
		if err != nil {
			return fmt.Errorf("load deputy uids: %w", err)
		}

		baseURL := resolveBaseURL(ctx, cfg.BaseURL)
		urls := make([]sitemapURL, 0, len(uids))
		for _, uid := range uids {
			urls = append(urls, sitemapURL{Loc: baseURL + "/deputes/" + url.PathEscape(uid)})
		}

		sm := urlsetXML{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		}

		ctx.Response().Header().Set(echo.HeaderContentType, "text/xml; charset=utf-8")
		return ctx.XML(http.StatusOK, sm)
	}
}

func (s *Server) SitemapGroups(cfg SitemapConfig) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uids, err := s.store.SitemapGroupUIDs(ctx.Request().Context())
		if err != nil {
			return fmt.Errorf("load group uids: %w", err)
		}

		baseURL := resolveBaseURL(ctx, cfg.BaseURL)
		urls := make([]sitemapURL, 0, len(uids))
		for _, uid := range uids {
			urls = append(urls, sitemapURL{Loc: baseURL + "/groupes/" + url.PathEscape(uid)})
		}

		sm := urlsetXML{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		}

		ctx.Response().Header().Set(echo.HeaderContentType, "text/xml; charset=utf-8")
		return ctx.XML(http.StatusOK, sm)
	}
}

func resolveBaseURL(ctx echo.Context, configured string) string {
	if configured != "" {
		return configured
	}
	return ctx.Scheme() + "://" + ctx.Request().Host
}

type SitemapConfig struct {
	BaseURL string
}
