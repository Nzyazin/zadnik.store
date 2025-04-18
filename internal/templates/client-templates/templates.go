package client_templates

import (
	"embed"
	"text/template"
	"io"
)

//go:embed templates/**/*
var files embed.FS

type BaseParams struct {
	Title string
	Description string
	View string
}

type IndexParams struct {
	BaseParams
	Error string
	Products []Product
}

type DeliveryParams struct {
	BaseParams
	Error string
}

type PaymentParams struct {
	BaseParams
	Error string
}

type GuaranteeParams struct {
	BaseParams
	Error string
}

type PolicyParams struct {
	BaseParams
}

type ThankParams struct {
	BaseParams
	Name string
}

type ErrorParams struct {
	BaseParams
	Message string
}

type TemplateFunctions struct {
	StaticWithHash func(string) string
}

type Templates struct {
	index *template.Template
	delivery *template.Template
	payment *template.Template
	guarantee *template.Template
	policy *template.Template
	thank *template.Template
	error *template.Template
	funcs template.FuncMap
}

func NewTemplates(tf TemplateFunctions) (*Templates, error) {
	t := &Templates{
		funcs: template.FuncMap{
			"staticWithHash": tf.StaticWithHash,
		},
	}

	if err := t.parseTemplates(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Templates) parseTemplates() error {
	baseTemplates := []string{
		"templates/layout/base.html",
		"templates/components/_layout/header.html",
		"templates/components/_layout/footer.html",
		"templates/components/_layout/cookies.html",
		"templates/components/_layout/meta.html",
	}

	indexTemplates := []string{
		"templates/pages/index.html",
		"templates/components/index/cap.html",
		"templates/components/index/products.html",
		"templates/components/index/showcase.html",
		"templates/components/index/production.html",
		"templates/components/index/order-steps.html",
		"templates/components/index/order-form.html",
		"templates/components/index/faq.html",
	}

	t.index = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, indexTemplates...)...),
	)

	deliveryTemplates := []string{
		"templates/pages/delivery.html",
		"templates/components/delivery/delivery.html",
	}

	t.delivery = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, deliveryTemplates...)...),
	)

	paymentTemplates := []string{
		"templates/pages/payment.html",
		"templates/components/payment/payment.html",
	}

	t.payment = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, paymentTemplates...)...),
	)

	guaranteeTemplates := []string{
		"templates/pages/guarantee.html",
		"templates/components/guarantee/guarantee.html",
	}

	t.guarantee = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, guaranteeTemplates...)...),
	)

	policyTemplates := []string{
		"templates/pages/policy.html",
		"templates/components/policy/policy.html",
	}

	t.policy = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, policyTemplates...)...),
	)

	thankTemplates := []string{
		"templates/pages/thank.html",
		"templates/components/thank/thank.html",
	}

	t.thank = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, thankTemplates...)...),
	)

	errorTemplates := []string{
		"templates/pages/error.html",
		"templates/components/error/error.html",
	}

	t.error = template.Must(
		template.New("base.html").
			Funcs(t.funcs).
			ParseFS(files, append(baseTemplates, errorTemplates...)...),
	)

	return nil
}

func (t *Templates) RenderIndex(w io.Writer, p IndexParams) error {
	p.View = "index"

	return t.index.Execute(w, p)
}

func (t *Templates) RenderDelivery(w io.Writer, p DeliveryParams) error {
	p.View = "delivery"

	return t.delivery.Execute(w, p)
}

func (t *Templates) RenderPayment(w io.Writer, p PaymentParams) error {
	p.View = "payment"

	return t.payment.Execute(w, p)
}

func (t *Templates) RenderGuarantee(w io.Writer, p GuaranteeParams) error {
	p.View = "guarantee"

	return t.guarantee.Execute(w, p)
}

func (t *Templates) RenderPolicy(w io.Writer, p PolicyParams) error {
	p.View = "policy"

	return t.policy.Execute(w, p)
}

func (t *Templates) RenderThank(w io.Writer, p ThankParams) error {
	p.View = "thank"

	return t.thank.Execute(w, p)
}

func (t *Templates) RenderError(w io.Writer, p ErrorParams) error {
	p.View = "error"

	return t.error.Execute(w, p)
}
