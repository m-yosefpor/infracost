package prices

import (
	"sync"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func PopulatePrices(resources []*schema.Resource) error {
	q := NewGraphQLQueryRunner()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		q.ReportSummary(resources)
	}()

	for _, r := range resources {
		if r.IsSkipped {
			continue
		}

		if err := GetPrices(r, q); err != nil {
			return err
		}
	}

	wg.Wait()

	return nil
}

func GetPrices(r *schema.Resource, q QueryRunner) error {
	results, err := q.RunQueries(r)
	if err != nil {
		return err
	}

	for _, r := range results {
		setCostComponentPrice(r.Resource, r.CostComponent, r.Result)
	}

	return nil
}

func setCostComponentPrice(r *schema.Resource, c *schema.CostComponent, res gjson.Result) {
	var p decimal.Decimal

	products := res.Get("data.products").Array()
	if len(products) == 0 {
		if c.IgnoreIfMissingPrice {
			log.Debugf("No products found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		log.Warnf("No products found for %s %s, using 0.00", r.Name, c.Name)
		c.SetPrice(decimal.Zero)
		return
	}
	if len(products) > 1 {
		log.Warnf("Multiple products found for %s %s, using the first product", r.Name, c.Name)
	}

	prices := products[0].Get("prices").Array()
	if len(prices) == 0 {
		if c.IgnoreIfMissingPrice {
			log.Debugf("No prices found for %s %s, ignoring since IgnoreIfMissingPrice is set.", r.Name, c.Name)
			r.RemoveCostComponent(c)
			return
		}

		log.Warnf("No prices found for %s %s, using 0.00", r.Name, c.Name)
		c.SetPrice(decimal.Zero)
		return
	}
	if len(prices) > 1 {
		log.Warnf("Multiple prices found for %s %s, using the first price", r.Name, c.Name)
	}

	var err error
	p, err = decimal.NewFromString(prices[0].Get("USD").String())
	if err != nil {
		log.Warnf("Error converting price (using 0.00) '%v': %s", prices[0].Get("USD").String(), err.Error())
		c.SetPrice(decimal.Zero)
		return
	}

	c.SetPrice(p)
	c.SetPriceHash(prices[0].Get("priceHash").String())
}
