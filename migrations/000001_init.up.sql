CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);

COMMENT ON TABLE subscriptions IS 'Таблица для хранения информации о подписках пользователей';
COMMENT ON COLUMN subscriptions.id IS 'Уникальный идентификатор подписки';
COMMENT ON COLUMN subscriptions.service_name IS 'Название сервиса, предоставляющего подписку';
COMMENT ON COLUMN subscriptions.price IS 'Стоимость месячной подписки в рублях';
COMMENT ON COLUMN subscriptions.user_id IS 'ID пользователя в формате UUID';
COMMENT ON COLUMN subscriptions.start_date IS 'Дата начала подписки';
COMMENT ON COLUMN subscriptions.end_date IS 'Опциональная дата окончания подписки';
COMMENT ON COLUMN subscriptions.created_at IS 'Дата и время создания записи';
COMMENT ON COLUMN subscriptions.updated_at IS 'Дата и время последнего обновления записи';

