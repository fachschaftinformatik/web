-- +goose Up
-- Keep the legal revision date in the same transaction as contact changes.
-- +goose StatementBegin
CREATE TRIGGER trg_legal_person_revision AFTER UPDATE OF name,user_id ON people
WHEN (OLD.user_id IS NOT NEW.user_id OR (NEW.user_id IS NULL AND OLD.name IS NOT NEW.name))
AND EXISTS (
    SELECT 1 FROM responsibility_members rm JOIN organization o
    ON rm.responsibility_id=json_extract(o.data,'$.ChairResponsibility')
    WHERE rm.person_id=NEW.id
)
BEGIN
    UPDATE organization SET data=json_set(data,'$.LegalUpdated',strftime('%Y-%m-%d','now')) WHERE id=1;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_legal_account_revision AFTER UPDATE OF name ON users
WHEN OLD.name IS NOT NEW.name AND EXISTS (
    SELECT 1 FROM people p JOIN responsibility_members rm ON rm.person_id=p.id
    JOIN organization o ON rm.responsibility_id=json_extract(o.data,'$.ChairResponsibility')
    WHERE p.user_id=NEW.id
)
BEGIN
    UPDATE organization SET data=json_set(data,'$.LegalUpdated',strftime('%Y-%m-%d','now')) WHERE id=1;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_legal_responsibility_revision AFTER UPDATE OF email,group_name ON responsibilities
WHEN (OLD.email IS NOT NEW.email OR OLD.group_name IS NOT NEW.group_name)
AND NEW.id=(SELECT json_extract(data,'$.ChairResponsibility') FROM organization WHERE id=1)
BEGIN
    UPDATE organization SET data=json_set(data,'$.LegalUpdated',strftime('%Y-%m-%d','now')) WHERE id=1;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_legal_assignment_added AFTER INSERT ON responsibility_members
WHEN NEW.responsibility_id=(SELECT json_extract(data,'$.ChairResponsibility') FROM organization WHERE id=1)
BEGIN
    UPDATE organization SET data=json_set(data,'$.LegalUpdated',strftime('%Y-%m-%d','now')) WHERE id=1;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_legal_assignment_removed AFTER DELETE ON responsibility_members
WHEN OLD.responsibility_id=(SELECT json_extract(data,'$.ChairResponsibility') FROM organization WHERE id=1)
BEGIN
    UPDATE organization SET data=json_set(data,'$.LegalUpdated',strftime('%Y-%m-%d','now')) WHERE id=1;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER trg_legal_assignment_removed;
DROP TRIGGER trg_legal_assignment_added;
DROP TRIGGER trg_legal_responsibility_revision;
DROP TRIGGER trg_legal_account_revision;
DROP TRIGGER trg_legal_person_revision;
