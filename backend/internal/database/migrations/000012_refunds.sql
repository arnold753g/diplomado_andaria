ALTER TABLE tour_package_departures
    ADD COLUMN minimum_review_at TIMESTAMPTZ,
    ADD COLUMN minimum_reviewed_at TIMESTAMPTZ,
    ADD COLUMN minimum_review_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

DO $$
DECLARE
    item RECORD;
BEGIN
    FOR item IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'tour_package_departures'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) ILIKE '%status%'
    LOOP
        EXECUTE format('ALTER TABLE tour_package_departures DROP CONSTRAINT %I', item.conname);
    END LOOP;
END $$;

ALTER TABLE tour_package_departures
    ADD CONSTRAINT tour_package_departures_status_allowed
    CHECK (status IN ('draft','open','confirmed','closed','minimum_review','cancelled','completed')),
    ADD CONSTRAINT tour_package_departures_cancelled_reason
    CHECK (status <> 'cancelled' OR length(trim(cancellation_reason)) > 0),
    ADD CONSTRAINT tour_package_departures_cancelled_at_check
    CHECK (status <> 'cancelled' OR cancelled_at IS NOT NULL),
    ADD CONSTRAINT tour_package_departures_minimum_review_check
    CHECK (status <> 'minimum_review' OR minimum_review_at IS NOT NULL);

ALTER TABLE tour_package_purchases
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN cancellation_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN refund_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN refund_requested_at TIMESTAMPTZ,
    ADD COLUMN refund_due_at TIMESTAMPTZ,
    ADD COLUMN refund_method VARCHAR(20),
    ADD COLUMN refund_qr BYTEA,
    ADD COLUMN refund_bank_name VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN refund_account_holder VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN refund_account_number VARCHAR(80) NOT NULL DEFAULT '',
    ADD COLUMN refund_proof BYTEA,
    ADD COLUMN refund_reference VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN refunded_at TIMESTAMPTZ,
    ADD COLUMN refund_completed_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;

UPDATE tour_package_purchases
SET refund_reason = COALESCE(NULLIF(trim(rejection_reason), ''), 'Reembolso pendiente'),
    refund_requested_at = COALESCE(reviewed_at, updated_at),
    refund_due_at = COALESCE(reviewed_at, updated_at) + INTERVAL '72 hours'
WHERE status IN ('refund_pending','refunded');

ALTER TABLE tour_package_purchases
    DROP CONSTRAINT IF EXISTS tour_package_purchases_review_state;

ALTER TABLE tour_package_purchases
    ADD CONSTRAINT tour_package_purchases_review_state
    CHECK (
        (status = 'payment_review' AND reviewed_at IS NULL AND rejection_reason = '') OR
        (status = 'correction_requested' AND reviewed_at IS NOT NULL AND length(trim(rejection_reason)) >= 5) OR
        (status = 'confirmed' AND reviewed_at IS NOT NULL AND rejection_reason = '') OR
        (status = 'payment_rejected' AND reviewed_at IS NOT NULL AND length(trim(rejection_reason)) >= 5) OR
        (status = 'refund_pending' AND refund_requested_at IS NOT NULL AND refund_due_at IS NOT NULL AND length(trim(refund_reason)) >= 5) OR
        (status = 'refunded' AND refund_requested_at IS NOT NULL AND refund_due_at IS NOT NULL AND refunded_at IS NOT NULL AND refund_proof IS NOT NULL) OR
        status = 'cancelled'
    ),
    ADD CONSTRAINT tour_package_purchases_refund_method_allowed
    CHECK (refund_method IS NULL OR refund_method IN ('','qr','bank_transfer')),
    ADD CONSTRAINT tour_package_purchases_refund_destination
    CHECK (
        refund_method IS NULL OR refund_method = '' OR
        (refund_method = 'qr' AND refund_qr IS NOT NULL AND octet_length(refund_qr) BETWEEN 1 AND 5242880) OR
        (refund_method = 'bank_transfer' AND length(trim(refund_bank_name)) >= 2 AND length(trim(refund_account_holder)) >= 3 AND length(trim(refund_account_number)) >= 3)
    ),
    ADD CONSTRAINT tour_package_purchases_refund_due_after_request
    CHECK (refund_due_at IS NULL OR (refund_requested_at IS NOT NULL AND refund_due_at > refund_requested_at)),
    ADD CONSTRAINT tour_package_purchases_refund_proof_size
    CHECK (refund_proof IS NULL OR octet_length(refund_proof) BETWEEN 1 AND 5242880),
    ADD CONSTRAINT tour_package_purchases_refunded_destination
    CHECK (status <> 'refunded' OR refund_method IN ('qr','bank_transfer'));

CREATE INDEX tour_package_purchases_refund_queue
    ON tour_package_purchases(agency_id, refund_due_at)
    WHERE status = 'refund_pending';

CREATE INDEX tour_package_departures_minimum_review
    ON tour_package_departures(booking_closes_at, starts_at)
    WHERE status IN ('open','confirmed','minimum_review');
