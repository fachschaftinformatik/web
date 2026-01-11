-- +goose Up
-- +goose StatementBegin

-- 1. Add Medieninformatik (B. Sc.) if it doesn't exist
INSERT OR IGNORE INTO programs (name) VALUES ('Medieninformatik (B. Sc.)');

-- 2. Link PO Versions for Medieninformatik (B. Sc.)
INSERT OR IGNORE INTO program_versions (programid, name)
SELECT id, 'PO2016' FROM programs WHERE name = 'Medieninformatik (B. Sc.)';
INSERT OR IGNORE INTO program_versions (programid, name)
SELECT id, 'PO2023' FROM programs WHERE name = 'Medieninformatik (B. Sc.)';

-- 3. Reconcile Modules & Aliases (Filtering out non-exam modules)
-- Using a temporary table to handle the mass updates/inserts cleanly
CREATE TEMP TABLE module_sync (
    program_name TEXT,
    module_name TEXT,
    module_alias TEXT
);

-- DATA FROM JSON (PO 2016 & 2023 combined, filtering out non-exams as we go)
-- Note: Reconciling specific variants like "Grundlagen der Mathematik für Informatiker"
INSERT INTO module_sync VALUES
-- Informatik (B. Sc.)
('Informatik (B. Sc.)', 'Algorithmen und Datenstrukturen', 'ADS'),
('Informatik (B. Sc.)', 'Betriebssysteme', 'BSY'),
('Informatik (B. Sc.)', 'Bildverarbeitung', 'BV'),
('Informatik (B. Sc.)', 'Datenbanksysteme', 'DBA'),
('Informatik (B. Sc.)', 'Echtzeitsysteme', 'EZS'),
('Informatik (B. Sc.)', 'Einführung in die Programmierung', 'EPR'),
('Informatik (B. Sc.)', 'Grundlagen der Mathematik für Informatiker', 'GMI'),
('Informatik (B. Sc.)', 'Internet-Datenbanken', 'IDB'),
('Informatik (B. Sc.)', 'Internet-Protokolle', 'INP'),
('Informatik (B. Sc.)', 'Internet-Sprachen', 'INS'),
('Informatik (B. Sc.)', 'Logik und diskrete Strukturen', 'LDS'),
('Informatik (B. Sc.)', 'Mathematik für Informatiker', 'MIN'),
('Informatik (B. Sc.)', 'Mathematische Grundlagen', 'MGR'),
('Informatik (B. Sc.)', 'Mensch-Computer-Interaktion', 'MCI'),
('Informatik (B. Sc.)', 'Objektorientierte Programmierung', 'OPR'),
('Informatik (B. Sc.)', 'Prozedurale Programmierung', 'PPR'),
('Informatik (B. Sc.)', 'Rechnernetze', 'REN'),
('Informatik (B. Sc.)', 'Robotik', 'ROB'),
('Informatik (B. Sc.)', 'Softwaretechnik', 'SWT'),
('Informatik (B. Sc.)', 'Technische Grundlagen der Informatik', 'TGI'),
('Informatik (B. Sc.)', 'Technisches Englisch für Informatiker', 'TENI'),
('Informatik (B. Sc.)', 'Theoretische Informatik', 'THI'),
('Informatik (B. Sc.)', 'Statistik und Lineare Algebra', 'SLA'),

-- Medieninformatik (B. Sc.)
('Medieninformatik (B. Sc.)', '3D-Computergrafik', 'CGR'),
('Medieninformatik (B. Sc.)', '3D-Modellierung und Animation', 'ANI'),
('Medieninformatik (B. Sc.)', 'Algorithmen und Datenstrukturen', 'ADS'),
('Medieninformatik (B. Sc.)', 'Betriebssysteme', 'BSY'),
('Medieninformatik (B. Sc.)', 'Bildgestaltung', 'BGS'),
('Medieninformatik (B. Sc.)', 'Datenbanksysteme', 'DBA'),
('Medieninformatik (B. Sc.)', 'Designgrundlagen', 'DSG'),
('Medieninformatik (B. Sc.)', 'Einführung in die Programmierung', 'EPR'),
('Medieninformatik (B. Sc.)', 'Grundlagen der Mathematik für Informatiker', 'GMI'),
('Medieninformatik (B. Sc.)', 'Internet-Sprachen', 'INS'),
('Medieninformatik (B. Sc.)', 'Logik und diskrete Strukturen', 'LDS'),
('Medieninformatik (B. Sc.)', 'Mathematik für Medieninformatiker', 'MMI'),
('Medieninformatik (B. Sc.)', 'Medientechnik', 'MET'),
('Medieninformatik (B. Sc.)', 'Mensch-Computer-Interaktion in der Medieninformatik', 'MCIM'),
('Medieninformatik (B. Sc.)', 'Objektorientierte Programmierung', 'OPR'),
('Medieninformatik (B. Sc.)', 'Softwaretechnik', 'SWT'),
('Medieninformatik (B. Sc.)', 'Technische Grundlagen der Informatik', 'TGI'),
('Medieninformatik (B. Sc.)', 'Technisches Englisch für Medieninformatiker', 'TENM'),
('Medieninformatik (B. Sc.)', 'Theoretische Informatik', 'THI'),

-- Wirtschaftsinformatik (B. Sc.)
('Wirtschaftsinformatik (B. Sc.)', 'Algorithmen und Datenstrukturen', 'ADS'),
('Wirtschaftsinformatik (B. Sc.)', 'Betriebliche Informationssysteme 1', 'BI1'),
('Wirtschaftsinformatik (B. Sc.)', 'Betriebliche Informationssysteme 2', 'BI2'),
('Wirtschaftsinformatik (B. Sc.)', 'Betriebliches Rechnungswesen', 'BRW'),
('Wirtschaftsinformatik (B. Sc.)', 'Betriebssysteme und Netzwerke für WI', 'BNW'),
('Wirtschaftsinformatik (B. Sc.)', 'Datenbanksysteme', 'DBA'),
('Wirtschaftsinformatik (B. Sc.)', 'Digitales Marketing', 'DIM'),
('Wirtschaftsinformatik (B. Sc.)', 'Einführung in die Betriebswirtschaftslehre', 'EBWL'),
('Wirtschaftsinformatik (B. Sc.)', 'Einführung in die Programmierung', 'EPR'),
('Wirtschaftsinformatik (B. Sc.)', 'Geschäftsprozessmanagement', 'GPM'),
('Wirtschaftsinformatik (B. Sc.)', 'Grundlagen der Mathematik für Informatiker', 'GMI'),
('Wirtschaftsinformatik (B. Sc.)', 'Grundlagen der Wirtschaftsinformatik', 'GWI'),
('Wirtschaftsinformatik (B. Sc.)', 'IT-Recht', 'ITR'),
('Wirtschaftsinformatik (B. Sc.)', 'Logik und diskrete Strukturen', 'LDS'),
('Wirtschaftsinformatik (B. Sc.)', 'Mathematik für Wirtschaftsinformatiker', 'MWI'),
('Wirtschaftsinformatik (B. Sc.)', 'Mensch-Computer-Interaktion in der Wirtschaftsinformatik', 'MCIW'),
('Wirtschaftsinformatik (B. Sc.)', 'Objektorientierte Programmierung', 'OPR'),
('Wirtschaftsinformatik (B. Sc.)', 'Produktion und Materialwirtschaft', 'PMW'),
('Wirtschaftsinformatik (B. Sc.)', 'Projektmanagement', 'PMA'),
('Wirtschaftsinformatik (B. Sc.)', 'Softwaretechnik', 'SWT'),
('Wirtschaftsinformatik (B. Sc.)', 'Wirtschaftsenglisch für Wirtschaftsinformatiker', 'WENW'),
('Wirtschaftsinformatik (B. Sc.)', 'Grundlagen Supply Chain Management', 'GSCM'),
('Wirtschaftsinformatik (B. Sc.)', 'Statistik und Lineare Algebra', 'SLA'),
('Wirtschaftsinformatik (B. Sc.)', 'Supply Chain Management und Digitalisierung', 'SCMD'),
('Wirtschaftsinformatik (B. Sc.)', 'Mathematische Grundlagen', 'MGR'),
('Wirtschaftsinformatik (B. Sc.)', 'Mensch-Computer-Interaktion', 'MCI'),

-- Informatik und Design (B. Sc.)
('Informatik und Design (B. Sc.)', 'Algorithmen und Datenstrukturen', 'ADS'),
('Informatik und Design (B. Sc.)', 'Cross-Platform Development', 'CPD'),
('Informatik und Design (B. Sc.)', 'Datenbanksysteme', 'DBA'),
('Informatik und Design (B. Sc.)', 'Einführung in die Programmierung', 'EPR'),
('Informatik und Design (B. Sc.)', 'Extended Reality', 'EXR'),
('Informatik und Design (B. Sc.)', 'Informatik und Design in Kultur und Gesellschaft', 'IDKG'),
('Informatik und Design (B. Sc.)', 'Logik und diskrete Strukturen', 'LDS'),
('Informatik und Design (B. Sc.)', 'Mensch-Computer-Interaktion', 'MCI'),
('Informatik und Design (B. Sc.)', 'Mathematische Grundlagen', 'MGR'),
('Informatik und Design (B. Sc.)', 'Objektorientierte Programmierung', 'OPR'),
('Informatik und Design (B. Sc.)', 'PRIMER to Building Sustainable Futures', 'PRB'),
('Informatik und Design (B. Sc.)', 'PRIMER to Designing Sustainable Futures', 'PRD'),
('Informatik und Design (B. Sc.)', 'Statistik und Lineare Algebra', 'SLA'),
('Informatik und Design (B. Sc.)', 'Projekt-Support-Modul BUILDING Sustainable Futures', 'SPB'),
('Informatik und Design (B. Sc.)', 'Projekt-Support-Modul DESIGNING Sustainable Futures', 'SPD'),
('Informatik und Design (B. Sc.)', 'START Design', 'SDES'),
('Informatik und Design (B. Sc.)', 'START Informatik', 'SINF'),
('Informatik und Design (B. Sc.)', 'Softwaretechnik', 'SWT'),
('Informatik und Design (B. Sc.)', 'Technisches Englisch', 'TEID'),

-- Informatik (M. Sc.)
('Informatik (M. Sc.)', 'Informatik und Gesellschaft', 'IGE'),
('Informatik (M. Sc.)', 'Master-Seminar Informatik', 'MSIN'),
('Informatik (M. Sc.)', 'Projektmanagement', 'PM'),
('Informatik (M. Sc.)', 'Wissenschaftliche Vertiefung Informatik', 'WVIN');

-- APPLY SYNC
-- A. Insert missing modules
INSERT OR IGNORE INTO modules (programid, name, alias)
SELECT p.id, s.module_name, s.module_alias
FROM module_sync s
JOIN programs p ON p.name = s.program_name;

-- B. Update aliases for existing modules
UPDATE modules
SET alias = (
    SELECT s.module_alias 
    FROM module_sync s 
    JOIN programs p ON p.name = s.program_name
    WHERE s.module_name = modules.name 
      AND p.id = modules.programid
)
WHERE EXISTS (
    SELECT 1 
    FROM module_sync s 
    JOIN programs p ON p.name = s.program_name
    WHERE s.module_name = modules.name 
      AND p.id = modules.programid
);

DROP TABLE module_sync;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Omitted as it would involve deleting a specific subset of data.
-- +goose StatementEnd
