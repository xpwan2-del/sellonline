package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestResolveOrderAffiliateSnapshotPreferLatestVisitorClick(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)

	promoterA := createAffiliateTestUser(t, db, "affiliate-a@example.com")
	promoterB := createAffiliateTestUser(t, db, "affiliate-b@example.com")
	profileA := createAffiliateTestProfile(t, db, promoterA.ID, "AFFA0001", constants.AffiliateProfileStatusActive)
	profileB := createAffiliateTestProfile(t, db, promoterB.ID, "AFFB0002", constants.AffiliateProfileStatusActive)

	visitorKey := "visitor-key-priority"
	now := time.Now()
	createAffiliateTestClick(t, db, profileA.ID, visitorKey, now.Add(-2*time.Hour))
	createAffiliateTestClick(t, db, profileB.ID, visitorKey, now.Add(-1*time.Hour))

	profileID, code, err := svc.ResolveOrderAffiliateSnapshot(0, profileA.AffiliateCode, visitorKey)
	if err != nil {
		t.Fatalf("resolve snapshot failed: %v", err)
	}
	if profileID == nil || *profileID != profileB.ID {
		t.Fatalf("expected latest clicked profile %d, got %+v", profileB.ID, profileID)
	}
	if code != profileB.AffiliateCode {
		t.Fatalf("expected latest clicked code %s, got %s", profileB.AffiliateCode, code)
	}
}

func TestResolveOrderAffiliateSnapshotFallbackToCodeWhenNoVisitorClick(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)

	promoter := createAffiliateTestUser(t, db, "affiliate-fallback@example.com")
	profile := createAffiliateTestProfile(t, db, promoter.ID, "AFFF0003", constants.AffiliateProfileStatusActive)

	profileID, code, err := svc.ResolveOrderAffiliateSnapshot(0, profile.AffiliateCode, "visitor-key-not-found")
	if err != nil {
		t.Fatalf("resolve snapshot failed: %v", err)
	}
	if profileID == nil || *profileID != profile.ID {
		t.Fatalf("expected fallback profile %d, got %+v", profile.ID, profileID)
	}
	if code != profile.AffiliateCode {
		t.Fatalf("expected fallback code %s, got %s", profile.AffiliateCode, code)
	}
}

func TestResolveOrderAffiliateSnapshotRejectSelfByVisitorClick(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)

	promoter := createAffiliateTestUser(t, db, "affiliate-self@example.com")
	profile := createAffiliateTestProfile(t, db, promoter.ID, "AFFS0004", constants.AffiliateProfileStatusActive)
	createAffiliateTestClick(t, db, profile.ID, "visitor-key-self", time.Now().Add(-10*time.Minute))

	profileID, code, err := svc.ResolveOrderAffiliateSnapshot(promoter.ID, "AFFF9999", "visitor-key-self")
	if err != nil {
		t.Fatalf("resolve snapshot failed: %v", err)
	}
	if profileID != nil || code != "" {
		t.Fatalf("expected self-order attribution ignored, got profile=%+v code=%q", profileID, code)
	}
}

func TestUpdateAffiliateProfileStatus(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)

	user := createAffiliateTestUser(t, db, "affiliate-status@example.com")
	profile := createAffiliateTestProfile(t, db, user.ID, "AFFST001", constants.AffiliateProfileStatusActive)

	disabled, err := svc.UpdateAffiliateProfileStatus(profile.ID, constants.AffiliateProfileStatusDisabled)
	if err != nil {
		t.Fatalf("disable profile failed: %v", err)
	}
	if disabled == nil || disabled.Status != constants.AffiliateProfileStatusDisabled {
		t.Fatalf("expected disabled status, got %+v", disabled)
	}

	enabled, err := svc.UpdateAffiliateProfileStatus(profile.ID, constants.AffiliateProfileStatusActive)
	if err != nil {
		t.Fatalf("enable profile failed: %v", err)
	}
	if enabled == nil || enabled.Status != constants.AffiliateProfileStatusActive {
		t.Fatalf("expected active status, got %+v", enabled)
	}
}

func TestBatchUpdateAffiliateProfileStatus(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)

	userA := createAffiliateTestUser(t, db, "affiliate-batch-a@example.com")
	userB := createAffiliateTestUser(t, db, "affiliate-batch-b@example.com")
	profileA := createAffiliateTestProfile(t, db, userA.ID, "AFFBT001", constants.AffiliateProfileStatusActive)
	profileB := createAffiliateTestProfile(t, db, userB.ID, "AFFBT002", constants.AffiliateProfileStatusActive)

	updated, err := svc.BatchUpdateAffiliateProfileStatus([]uint{profileA.ID, profileB.ID}, constants.AffiliateProfileStatusDisabled)
	if err != nil {
		t.Fatalf("batch disable failed: %v", err)
	}
	if updated != 2 {
		t.Fatalf("expected updated 2, got %d", updated)
	}

	reloadedA, err := svc.repo.GetProfileByID(profileA.ID)
	if err != nil || reloadedA == nil {
		t.Fatalf("reload profileA failed: %v", err)
	}
	reloadedB, err := svc.repo.GetProfileByID(profileB.ID)
	if err != nil || reloadedB == nil {
		t.Fatalf("reload profileB failed: %v", err)
	}
	if reloadedA.Status != constants.AffiliateProfileStatusDisabled || reloadedB.Status != constants.AffiliateProfileStatusDisabled {
		t.Fatalf("unexpected statuses after batch disable: %s, %s", reloadedA.Status, reloadedB.Status)
	}
}

func TestCreateAgentApplicationRequiresPhone(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)
	user := createAffiliateTestUser(t, db, "agent-application-phone@example.com")

	_, err := svc.CreateAgentApplication(AffiliateAgentApplicationCreateInput{
		UserID: user.ID,
		Phone:  "",
	})
	if !errors.Is(err, ErrAffiliateAgentApplicationInvalid) {
		t.Fatalf("expected invalid application error, got %v", err)
	}
}

func TestCreateAgentApplicationCreatesPendingApplication(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)
	user := createAffiliateTestUser(t, db, "agent-application-create@example.com")

	row, err := svc.CreateAgentApplication(AffiliateAgentApplicationCreateInput{
		UserID:  user.ID,
		Phone:   "13800138000",
		Message: "I want to promote products",
	})
	if err != nil {
		t.Fatalf("create agent application failed: %v", err)
	}
	if row == nil || row.ID == 0 {
		t.Fatalf("expected created application, got %+v", row)
	}
	if row.UserID != user.ID || row.Email != user.Email {
		t.Fatalf("unexpected application user info: %+v", row)
	}
	if row.Phone != "13800138000" {
		t.Fatalf("expected phone saved, got %q", row.Phone)
	}
	if row.Status != constants.AffiliateAgentApplicationStatusPending {
		t.Fatalf("expected pending status, got %s", row.Status)
	}
}

func TestCreateAgentApplicationRejectsDuplicatePending(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)
	user := createAffiliateTestUser(t, db, "agent-application-duplicate@example.com")

	if _, err := svc.CreateAgentApplication(AffiliateAgentApplicationCreateInput{
		UserID: user.ID,
		Phone:  "13800138000",
	}); err != nil {
		t.Fatalf("create first application failed: %v", err)
	}
	_, err := svc.CreateAgentApplication(AffiliateAgentApplicationCreateInput{
		UserID: user.ID,
		Phone:  "13800138001",
	})
	if !errors.Is(err, ErrAffiliateAgentApplicationExists) {
		t.Fatalf("expected duplicate pending error, got %v", err)
	}
}

func TestCreateAgentApplicationRejectsExistingAffiliate(t *testing.T) {
	svc, db := setupAffiliateServiceTest(t)
	user := createAffiliateTestUser(t, db, "agent-application-opened@example.com")
	createAffiliateTestProfile(t, db, user.ID, "AFFOPEN1", constants.AffiliateProfileStatusActive)

	_, err := svc.CreateAgentApplication(AffiliateAgentApplicationCreateInput{
		UserID: user.ID,
		Phone:  "13800138000",
	})
	if !errors.Is(err, ErrAffiliateAgentApplicationInvalid) {
		t.Fatalf("expected invalid application error, got %v", err)
	}
}

func TestHandleOrderPaidCreatesProductLevelCommissionItems(t *testing.T) {
	svc, db := setupAffiliateCommissionServiceTest(t, true, 30)

	promoter := createAffiliateTestUser(t, db, "affiliate-mixed-promoter@example.com")
	profile := createAffiliateTestProfile(t, db, promoter.ID, "AFFMIX01", constants.AffiliateProfileStatusActive)
	customer := createAffiliateTestUser(t, db, "affiliate-mixed-customer@example.com")

	rate5 := models.NewMoneyFromDecimal(decimal.NewFromInt(5))
	rate15 := models.NewMoneyFromDecimal(decimal.NewFromInt(15))
	productA := createAffiliateTestProduct(t, db, "affiliate-mixed-a", true, &rate5)
	productB := createAffiliateTestProduct(t, db, "affiliate-mixed-b", true, &rate15)
	productC := createAffiliateTestProduct(t, db, "affiliate-mixed-c", true, nil)

	order := createAffiliateTestOrder(t, db, customer.ID, profile.ID, profile.AffiliateCode, []models.OrderItem{
		createAffiliateTestOrderItem(productA.ID, "商品A", "100.00", "0.00", 3),
		createAffiliateTestOrderItem(productB.ID, "商品B", "200.00", "0.00", 1),
		createAffiliateTestOrderItem(productC.ID, "商品C", "300.00", "0.00", 10),
	})

	if err := svc.HandleOrderPaid(order.ID); err != nil {
		t.Fatalf("handle order paid failed: %v", err)
	}

	var commission models.AffiliateCommission
	if err := db.Preload("Items").Where("order_id = ?", order.ID).First(&commission).Error; err != nil {
		t.Fatalf("load commission failed: %v", err)
	}
	assertDecimalString(t, commission.BaseAmount.Decimal, "600.00")
	assertDecimalString(t, commission.CommissionAmount.Decimal, "125.00")
	assertDecimalString(t, commission.RatePercent.Decimal, "20.83")
	if len(commission.Items) != 3 {
		t.Fatalf("expected 3 commission items, got %d", len(commission.Items))
	}

	byProduct := map[uint]models.AffiliateCommissionItem{}
	for _, item := range commission.Items {
		byProduct[item.ProductID] = item
	}
	assertDecimalString(t, byProduct[productA.ID].RatePercent.Decimal, "5.00")
	assertDecimalString(t, byProduct[productA.ID].CommissionAmount.Decimal, "5.00")
	assertDecimalString(t, byProduct[productB.ID].RatePercent.Decimal, "15.00")
	assertDecimalString(t, byProduct[productB.ID].CommissionAmount.Decimal, "30.00")
	assertDecimalString(t, byProduct[productC.ID].RatePercent.Decimal, "30.00")
	assertDecimalString(t, byProduct[productC.ID].CommissionAmount.Decimal, "90.00")
}

func TestHandleOrderPaidUsesProductRateWhenPlatformDefaultDisabled(t *testing.T) {
	svc, db := setupAffiliateCommissionServiceTest(t, false, 30)

	promoter := createAffiliateTestUser(t, db, "affiliate-custom-promoter@example.com")
	profile := createAffiliateTestProfile(t, db, promoter.ID, "AFFCUS01", constants.AffiliateProfileStatusActive)
	customer := createAffiliateTestUser(t, db, "affiliate-custom-customer@example.com")

	rate10 := models.NewMoneyFromDecimal(decimal.NewFromInt(10))
	customProduct := createAffiliateTestProduct(t, db, "affiliate-custom-rate", true, &rate10)
	defaultProduct := createAffiliateTestProduct(t, db, "affiliate-default-disabled", true, nil)

	order := createAffiliateTestOrder(t, db, customer.ID, profile.ID, profile.AffiliateCode, []models.OrderItem{
		createAffiliateTestOrderItem(customProduct.ID, "单独比例商品", "100.00", "0.00", 1),
		createAffiliateTestOrderItem(defaultProduct.ID, "平台默认商品", "200.00", "0.00", 1),
	})

	if err := svc.HandleOrderPaid(order.ID); err != nil {
		t.Fatalf("handle order paid failed: %v", err)
	}

	var commission models.AffiliateCommission
	if err := db.Preload("Items").Where("order_id = ?", order.ID).First(&commission).Error; err != nil {
		t.Fatalf("load commission failed: %v", err)
	}
	assertDecimalString(t, commission.BaseAmount.Decimal, "100.00")
	assertDecimalString(t, commission.CommissionAmount.Decimal, "10.00")
	assertDecimalString(t, commission.RatePercent.Decimal, "10.00")
	if len(commission.Items) != 1 {
		t.Fatalf("expected only custom product commission item, got %d", len(commission.Items))
	}
	if commission.Items[0].ProductID != customProduct.ID {
		t.Fatalf("expected custom product item, got product %d", commission.Items[0].ProductID)
	}
}

func TestListMyCustomersReturnsRowsWithLastOrderAt(t *testing.T) {
	svc, db := setupAffiliateCommissionServiceTest(t, true, 30)

	promoter := createAffiliateTestUser(t, db, "customer-list-promoter@example.com")
	customer := createAffiliateTestUser(t, db, "customer-list-buyer@example.com")
	profile := createAffiliateTestProfile(t, db, promoter.ID, "CUSTLIST", constants.AffiliateProfileStatusActive)
	product := createAffiliateTestProduct(t, db, "customer-list-product", true, nil)
	order := createAffiliateTestOrder(t, db, customer.ID, profile.ID, profile.AffiliateCode, []models.OrderItem{
		createAffiliateTestOrderItem(product.ID, "customer-list-product", "450.00", "0.00", 1),
	})
	commission := models.AffiliateCommission{
		AffiliateProfileID: profile.ID,
		OrderID:            order.ID,
		CommissionType:     constants.AffiliateCommissionTypeOrder,
		BaseAmount:         models.NewMoneyFromDecimal(decimal.NewFromInt(450)),
		RatePercent:        models.NewMoneyFromDecimal(decimal.NewFromInt(30)),
		CommissionAmount:   models.NewMoneyFromDecimal(decimal.NewFromInt(135)),
		Status:             constants.AffiliateCommissionStatusAvailable,
		AvailableAt:        order.PaidAt,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := db.Create(&commission).Error; err != nil {
		t.Fatalf("create commission failed: %v", err)
	}
	relation := models.AffiliateCustomerRelation{
		CustomerUserID:      customer.ID,
		AffiliateProfileID:  profile.ID,
		SourceAffiliateCode: profile.AffiliateCode,
		SourceType:          constants.AffiliateCustomerRelationSourceRegister,
		BoundAt:             time.Now(),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	if err := db.Create(&relation).Error; err != nil {
		t.Fatalf("create customer relation failed: %v", err)
	}

	rows, total, err := svc.ListMyCustomers(promoter.ID, AffiliateCustomerRelationListInput{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list my customers failed: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("expected one customer, total=%d rows=%d", total, len(rows))
	}
	if rows[0].Email != customer.Email {
		t.Fatalf("expected customer email %s, got %s", customer.Email, rows[0].Email)
	}
	if rows[0].LastOrderAt == nil {
		t.Fatalf("expected last order time")
	}
	if rows[0].OrderCount != 1 {
		t.Fatalf("expected order count 1, got %d", rows[0].OrderCount)
	}
	assertDecimalString(t, rows[0].CommissionAmount.Decimal, "135.00")
}

func TestApplyWithdrawAllowsAvailableCommissionWhenPlatformDefaultDisabled(t *testing.T) {
	svc, db := setupAffiliateCommissionServiceTest(t, false, 30)

	user := createAffiliateTestUser(t, db, "affiliate-withdraw-disabled-default@example.com")
	profile := createAffiliateTestProfile(t, db, user.ID, "AFFWDF01", constants.AffiliateProfileStatusActive)
	now := time.Now()
	commission := models.AffiliateCommission{
		AffiliateProfileID: profile.ID,
		OrderID:            1,
		CommissionType:     constants.AffiliateCommissionTypeOrder,
		BaseAmount:         models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		RatePercent:        models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		CommissionAmount:   models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		Status:             constants.AffiliateCommissionStatusAvailable,
		AvailableAt:        &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := db.Create(&commission).Error; err != nil {
		t.Fatalf("create commission failed: %v", err)
	}

	withdraw, err := svc.ApplyWithdraw(user.ID, AffiliateWithdrawApplyInput{
		Amount:  decimal.NewFromInt(10),
		Channel: "bank",
		Account: "account",
	})
	if err != nil {
		t.Fatalf("apply withdraw failed: %v", err)
	}
	if withdraw == nil || !withdraw.Amount.Decimal.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected withdraw amount 10, got %+v", withdraw)
	}
}

func TestHandleOrderRefundedTxReducesCommissionItemsAndRecomputesRate(t *testing.T) {
	svc, db := setupAffiliateCommissionServiceTest(t, true, 30)

	promoter := createAffiliateTestUser(t, db, "affiliate-refund-promoter@example.com")
	profile := createAffiliateTestProfile(t, db, promoter.ID, "AFFREF01", constants.AffiliateProfileStatusActive)
	customer := createAffiliateTestUser(t, db, "affiliate-refund-customer@example.com")

	rate10 := models.NewMoneyFromDecimal(decimal.NewFromInt(10))
	rate20 := models.NewMoneyFromDecimal(decimal.NewFromInt(20))
	productA := createAffiliateTestProduct(t, db, "affiliate-refund-a", true, &rate10)
	productB := createAffiliateTestProduct(t, db, "affiliate-refund-b", true, &rate20)
	order := createAffiliateTestOrder(t, db, customer.ID, profile.ID, profile.AffiliateCode, []models.OrderItem{
		createAffiliateTestOrderItem(productA.ID, "退款商品A", "100.00", "0.00", 1),
		createAffiliateTestOrderItem(productB.ID, "退款商品B", "100.00", "0.00", 1),
	})

	if err := svc.HandleOrderPaid(order.ID); err != nil {
		t.Fatalf("handle order paid failed: %v", err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		return svc.HandleOrderRefundedTx(tx, &order, decimal.NewFromInt(50), decimal.Zero, "partial_refund")
	}); err != nil {
		t.Fatalf("handle refund failed: %v", err)
	}

	var commission models.AffiliateCommission
	if err := db.Preload("Items").Where("order_id = ?", order.ID).First(&commission).Error; err != nil {
		t.Fatalf("load commission failed: %v", err)
	}
	assertDecimalString(t, commission.BaseAmount.Decimal, "150.00")
	assertDecimalString(t, commission.CommissionAmount.Decimal, "22.50")
	assertDecimalString(t, commission.RatePercent.Decimal, "15.00")

	byProduct := map[uint]models.AffiliateCommissionItem{}
	for _, item := range commission.Items {
		byProduct[item.ProductID] = item
	}
	assertDecimalString(t, byProduct[productA.ID].BaseAmount.Decimal, "75.00")
	assertDecimalString(t, byProduct[productA.ID].CommissionAmount.Decimal, "7.50")
	assertDecimalString(t, byProduct[productB.ID].BaseAmount.Decimal, "75.00")
	assertDecimalString(t, byProduct[productB.ID].CommissionAmount.Decimal, "15.00")
}

func TestApplyWithdrawSplitsCommissionItemsAndRecomputesRates(t *testing.T) {
	svc, db := setupAffiliateCommissionServiceTest(t, false, 30)

	user := createAffiliateTestUser(t, db, "affiliate-withdraw-split@example.com")
	profile := createAffiliateTestProfile(t, db, user.ID, "AFFWSP01", constants.AffiliateProfileStatusActive)
	now := time.Now()
	commission := models.AffiliateCommission{
		AffiliateProfileID: profile.ID,
		OrderID:            99,
		CommissionType:     constants.AffiliateCommissionTypeOrder,
		BaseAmount:         models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		RatePercent:        models.NewMoneyFromDecimal(decimal.NewFromInt(20)),
		CommissionAmount:   models.NewMoneyFromDecimal(decimal.NewFromInt(20)),
		Status:             constants.AffiliateCommissionStatusAvailable,
		AvailableAt:        &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := db.Create(&commission).Error; err != nil {
		t.Fatalf("create commission failed: %v", err)
	}
	item := models.AffiliateCommissionItem{
		AffiliateCommissionID: commission.ID,
		OrderItemID:           99,
		ProductID:             99,
		BaseAmount:            models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		RatePercent:           models.NewMoneyFromDecimal(decimal.NewFromInt(20)),
		CommissionAmount:      models.NewMoneyFromDecimal(decimal.NewFromInt(20)),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("create commission item failed: %v", err)
	}

	withdraw, err := svc.ApplyWithdraw(user.ID, AffiliateWithdrawApplyInput{
		Amount:  decimal.NewFromInt(5),
		Channel: "bank",
		Account: "account",
	})
	if err != nil {
		t.Fatalf("apply withdraw failed: %v", err)
	}
	if withdraw == nil || !withdraw.Amount.Decimal.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("expected withdraw amount 5, got %+v", withdraw)
	}

	var commissions []models.AffiliateCommission
	if err := db.Preload("Items").Where("order_id = ?", commission.OrderID).Order("id asc").Find(&commissions).Error; err != nil {
		t.Fatalf("load split commissions failed: %v", err)
	}
	if len(commissions) != 2 {
		t.Fatalf("expected 2 split commissions, got %d", len(commissions))
	}
	assertDecimalString(t, commissions[0].BaseAmount.Decimal, "25.00")
	assertDecimalString(t, commissions[0].CommissionAmount.Decimal, "5.00")
	assertDecimalString(t, commissions[0].RatePercent.Decimal, "20.00")
	if len(commissions[0].Items) != 1 {
		t.Fatalf("expected bound commission item, got %d", len(commissions[0].Items))
	}
	assertDecimalString(t, commissions[0].Items[0].BaseAmount.Decimal, "25.00")
	assertDecimalString(t, commissions[0].Items[0].CommissionAmount.Decimal, "5.00")

	assertDecimalString(t, commissions[1].BaseAmount.Decimal, "75.00")
	assertDecimalString(t, commissions[1].CommissionAmount.Decimal, "15.00")
	assertDecimalString(t, commissions[1].RatePercent.Decimal, "20.00")
	if len(commissions[1].Items) != 1 {
		t.Fatalf("expected remaining commission item, got %d", len(commissions[1].Items))
	}
	assertDecimalString(t, commissions[1].Items[0].BaseAmount.Decimal, "75.00")
	assertDecimalString(t, commissions[1].Items[0].CommissionAmount.Decimal, "15.00")
}

func setupAffiliateServiceTest(t *testing.T) (*AffiliateService, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:affiliate_service_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.AffiliateProfile{},
		&models.AffiliateInviteCode{},
		&models.AffiliateAgentApplication{},
		&models.AffiliateCustomerRelation{},
		&models.AffiliateClick{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	settingRepo := newMockSettingRepo()
	settingSvc := NewSettingService(settingRepo)
	if _, err := settingSvc.UpdateAffiliateSetting(AffiliateSetting{
		Enabled:        true,
		CommissionRate: 20,
	}); err != nil {
		t.Fatalf("init affiliate setting failed: %v", err)
	}

	affiliateRepo := repository.NewAffiliateRepository(db)
	return NewAffiliateService(affiliateRepo, repository.NewUserRepository(db), nil, nil, settingSvc), db
}

func setupAffiliateCommissionServiceTest(t *testing.T, platformDefaultEnabled bool, platformRate float64) (*AffiliateService, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:affiliate_commission_service_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.AffiliateProfile{},
		&models.AffiliateInviteCode{},
		&models.AffiliateCustomerRelation{},
		&models.AffiliateClick{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Fulfillment{},
		&models.AffiliateCommission{},
		&models.AffiliateCommissionItem{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	settingRepo := newMockSettingRepo()
	settingSvc := NewSettingService(settingRepo)
	if _, err := settingSvc.UpdateAffiliateSetting(AffiliateSetting{
		Enabled:        platformDefaultEnabled,
		CommissionRate: platformRate,
	}); err != nil {
		t.Fatalf("init affiliate setting failed: %v", err)
	}

	affiliateRepo := repository.NewAffiliateRepository(db)
	return NewAffiliateService(
		affiliateRepo,
		repository.NewUserRepository(db),
		repository.NewOrderRepository(db),
		repository.NewProductRepository(db),
		settingSvc,
	), db
}

func createAffiliateTestUser(t *testing.T, db *gorm.DB, email string) models.User {
	t.Helper()

	row := models.User{
		Email:        email,
		PasswordHash: "hash",
		DisplayName:  "tester",
		Status:       constants.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	return row
}

func createAffiliateTestProfile(t *testing.T, db *gorm.DB, userID uint, code, status string) models.AffiliateProfile {
	t.Helper()

	row := models.AffiliateProfile{
		UserID:        userID,
		AffiliateCode: code,
		Status:        status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("create affiliate profile failed: %v", err)
	}
	return row
}

func createAffiliateTestClick(t *testing.T, db *gorm.DB, profileID uint, visitorKey string, createdAt time.Time) {
	t.Helper()

	row := models.AffiliateClick{
		AffiliateProfileID: profileID,
		VisitorKey:         visitorKey,
		LandingPath:        "/",
		CreatedAt:          createdAt,
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("create affiliate click failed: %v", err)
	}
}

func createAffiliateTestProduct(t *testing.T, db *gorm.DB, slug string, affiliateEnabled bool, rate *models.Money) models.Product {
	t.Helper()

	now := time.Now()
	row := models.Product{
		CategoryID:              1,
		Slug:                    slug,
		TitleJSON:               models.JSON{"zh-CN": slug},
		PriceAmount:             models.NewMoneyFromDecimal(decimal.NewFromInt(100)),
		CostPriceAmount:         models.NewMoneyFromDecimal(decimal.Zero),
		AffiliateCommissionRate: rate,
		PurchaseType:            constants.ProductPurchaseMember,
		FulfillmentType:         constants.FulfillmentTypeManual,
		ManualStockTotal:        -1,
		IsAffiliateEnabled:      affiliateEnabled,
		IsActive:                true,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	return row
}

func createAffiliateTestOrderItem(productID uint, title, total, coupon string, qty int) models.OrderItem {
	totalAmount := decimal.RequireFromString(total)
	couponAmount := decimal.RequireFromString(coupon)
	unitAmount := totalAmount
	if qty > 0 {
		unitAmount = totalAmount.Div(decimal.NewFromInt(int64(qty))).Round(2)
	}

	return models.OrderItem{
		ProductID:          productID,
		TitleJSON:          models.JSON{"zh-CN": title},
		OriginalUnitPrice:  models.NewMoneyFromDecimal(unitAmount),
		UnitPrice:          models.NewMoneyFromDecimal(unitAmount),
		CostPrice:          models.NewMoneyFromDecimal(decimal.Zero),
		Quantity:           qty,
		OriginalTotalPrice: models.NewMoneyFromDecimal(totalAmount),
		TotalPrice:         models.NewMoneyFromDecimal(totalAmount),
		CouponDiscount:     models.NewMoneyFromDecimal(couponAmount),
		FulfillmentType:    constants.FulfillmentTypeManual,
	}
}

func createAffiliateTestOrder(t *testing.T, db *gorm.DB, userID, profileID uint, code string, items []models.OrderItem) models.Order {
	t.Helper()

	now := time.Now()
	originalAmount := decimal.Zero
	discountAmount := decimal.Zero
	for _, item := range items {
		originalAmount = originalAmount.Add(item.TotalPrice.Decimal)
		discountAmount = discountAmount.Add(item.CouponDiscount.Decimal)
		discountAmount = discountAmount.Add(item.MemberDiscount.Decimal)
		discountAmount = discountAmount.Add(item.PromotionDiscount.Decimal)
		discountAmount = discountAmount.Add(item.WholesaleDiscount.Decimal)
	}
	totalAmount := originalAmount.Sub(discountAmount)
	if totalAmount.IsNegative() {
		totalAmount = decimal.Zero
	}

	order := models.Order{
		OrderNo:            fmt.Sprintf("AFFTEST%d", now.UnixNano()),
		UserID:             userID,
		Status:             constants.OrderStatusPaid,
		Currency:           "CNY",
		OriginalAmount:     models.NewMoneyFromDecimal(originalAmount),
		DiscountAmount:     models.NewMoneyFromDecimal(discountAmount),
		TotalAmount:        models.NewMoneyFromDecimal(totalAmount),
		OnlinePaidAmount:   models.NewMoneyFromDecimal(totalAmount),
		AffiliateProfileID: &profileID,
		AffiliateCode:      code,
		PaidAt:             &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	for i := range items {
		items[i].OrderID = order.ID
	}
	if len(items) > 0 {
		if err := db.Create(&items).Error; err != nil {
			t.Fatalf("create order items failed: %v", err)
		}
	}
	order.Items = items
	return order
}

func assertDecimalString(t *testing.T, got decimal.Decimal, want string) {
	t.Helper()

	if got.Round(2).StringFixed(2) != want {
		t.Fatalf("expected %s, got %s", want, got.Round(2).StringFixed(2))
	}
}
