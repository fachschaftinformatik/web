-- +goose Up
-- +goose StatementBegin
-- Wirtschaftsinformatik (B. Sc.)
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'Betriebliche Informationssysteme 1', 'BIS1' FROM programs WHERE name = 'Wirtschaftsinformatik (B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'Betriebliche Informationssysteme 2', 'BIS2' FROM programs WHERE name = 'Wirtschaftsinformatik (B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'Betriebssysteme und Netzwerke für WI', 'BSNW' FROM programs WHERE name = 'Wirtschaftsinformatik (B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'Grundlagen Supply Chain Management', 'GSCM' FROM programs WHERE name = 'Wirtschaftsinformatik (B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'Mathematik für Wirtschaftsinformatiker', 'MWI' FROM programs WHERE name = 'Wirtschaftsinformatik (B. Sc.)';

-- Informatik und Design (B. Sc.)
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'Cross-Platform Development', 'CPD' FROM programs WHERE name = 'Informatik und Design (B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'Extended Reality', 'EXR' FROM programs WHERE name = 'Informatik und Design (B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'START Design', 'SDES' FROM programs WHERE name = 'Informatik und Design (B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name, alias) 
SELECT id, 'START Informatik', 'SINF' FROM programs WHERE name = 'Informatik und Design (B. Sc.)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM modules WHERE alias IN ('BIS1', 'BIS2', 'BSNW', 'GSCM', 'MWI', 'CPD', 'EXR', 'SDES', 'SINF');
-- +goose StatementEnd
