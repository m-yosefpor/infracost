package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestEBSSnapshotCopy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ebs_volume" "gp2" {
			availability_zone = "us-east-1a"
			size              = 10
		}

		resource "aws_ebs_snapshot" "gp2" {
			volume_id = aws_ebs_volume.gp2.id
		}

		resource "aws_ebs_snapshot_copy" "gp2" {
			source_snapshot_id = aws_ebs_snapshot.gp2.id
			source_region      = "us-east-1"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_ebs_volume.gp2",
			SkipCheck: true,
		},
		{
			Name:      "aws_ebs_snapshot.gp2",
			SkipCheck: true,
		},
		{
			Name: "aws_ebs_snapshot_copy.gp2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "EBS snapshot storage",
					PriceHash:       "63a6765e67e0ebcd29f15f1570b5e692-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "Fast snapshot restore",
					PriceHash:       "c8e7cffde49d51c97e8ec2cfb97e4557-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "ListChangedBlocks & ListSnapshotBlocks API requests",
					PriceHash:       "c5e9f6869c2ca75ebfbf6d1b0fb99a16-4a9dfd3965ffcbab75845ead7a27fd47",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "GetSnapshotBlock API requests",
					PriceHash:       "7e9c5258c113e0c54f63e43889ade9a7-d41397dab24f1e4fcce3916e21c3cec4",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "PutSnapshotBlock API requests",
					PriceHash:       "16002a3a5d722ade9816ff144a7dd91a-d41397dab24f1e4fcce3916e21c3cec4",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
