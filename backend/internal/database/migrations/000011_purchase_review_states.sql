DO $$
DECLARE
    item RECORD;
BEGIN
    FOR item IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'tour_package_purchases'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) ILIKE '%status%'
    LOOP
        EXECUTE format('ALTER TABLE tour_package_purchases DROP CONSTRAINT %I', item.conname);
    END LOOP;
END $$;

ALTER TABLE tour_package_purchases
    ADD CONSTRAINT tour_package_purchases_status_allowed
    CHECK (status IN ('payment_review','correction_requested','confirmed','payment_rejected','cancelled','refund_pending','refunded')),
    ADD CONSTRAINT tour_package_purchases_review_state
    CHECK (
        (status = 'payment_review' AND reviewed_at IS NULL AND rejection_reason = '') OR
        (status = 'correction_requested' AND reviewed_at IS NOT NULL AND length(trim(rejection_reason)) >= 5) OR
        (status = 'confirmed' AND reviewed_at IS NOT NULL AND rejection_reason = '') OR
        (status = 'payment_rejected' AND reviewed_at IS NOT NULL AND length(trim(rejection_reason)) >= 5) OR
        (status = 'refund_pending' AND reviewed_at IS NOT NULL AND length(trim(rejection_reason)) >= 5) OR
        (status IN ('cancelled','refunded'))
    );
