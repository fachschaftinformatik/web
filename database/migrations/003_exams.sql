-- +goose Up
-- +goose StatementBegin
CREATE TABLE modules (
  id INTEGER PRIMARY KEY,
  programid INTEGER NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(programid, name)
) STRICT;

CREATE INDEX idx_modules_program ON modules(programid);

ALTER TABLE exams ADD COLUMN moduleid INTEGER REFERENCES modules(id) ON DELETE RESTRICT;
ALTER TABLE exams ADD COLUMN comment TEXT;

CREATE INDEX idx_exams_module ON exams(moduleid);

--INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Logik und diskrete Strukturen' FROM programs WHERE name LIKE 'Informatik%';
--INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in die Programmierung' FROM programs WHERE name LIKE 'Informatik%';
--INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Datenbanksysteme' FROM programs WHERE name LIKE 'Informatik%';
--INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mathematische Grundlagen' FROM programs WHERE name LIKE 'Informatik%';
--INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Theoretische Informatik' FROM programs WHERE name LIKE 'Informatik%';
--INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Rechnernetze' FROM programs WHERE name LIKE 'Informatik%';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Algorithmen und Datenstrukturen' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Bachelorarbeit Informatik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Betriebssysteme' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Datenbanksysteme' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in die Programmierung' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Internet-Sprachen' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Kolloquium zur Bachelorarbeit Informatik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Logik und diskrete Strukturen' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mensch-Computer-Interaktion' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mathematische Grundlagen' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Objektorientierte Programmierung' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Praxisphase' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Rechnernetze' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Statistik und Lineare Algebra' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Softwareprojekt Informatik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Softwaretechnik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Technisches Englisch für Informatiker' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Technische Grundlagen der Informatik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Theoretische Informatik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Betrieb komplexer verteilter Systeme' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in die Bildverarbeitung' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Data on the Web' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Data Science in Practice' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in Cloud Computing' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in die medizinische Informatik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in die Robotik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Internet-Protokolle' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'IT-Recht' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Grundlagen der IT Sicherheit' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Komponentenbasierte Softwareentwicklung' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Künstliche Intelligenz' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Knowledge Graphs' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mobile Application Development' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mobile und Cloud Computing' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mobile Robotik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Parallele Programmierung' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Prozedurale Programmierung' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Practical Security Attacks and Exploitation' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Software Design' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Betriebliches Rechnungswesen' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Digitales Marketing' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Einführung in die Betriebswirtschaftslehre' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Geschäftsprozessmanagement' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Grundlagen der Wirtschaftsinformatik' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Angewandte Netzwerksicherheit' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Projektmanagement' FROM programs WHERE name LIKE '%(B. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Produktion und Materialwirtschaft' FROM programs WHERE name LIKE '%(B. Sc.)';

INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Informatik und Gesellschaft' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Kolloquium zur Masterarbeit Informatik' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Masterarbeit Informatik' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Master-Projekt Informatik' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Master-Seminar Informatik' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Projektmanagement' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Wissenschaftliche Vertiefung Informatik' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Autonome Systeme' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Computer Vision' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Datenbanktheorie' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Data Science Principles' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Entwicklung intelligenter Systeme' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Future Computing' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Funktionale Programmierung' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Intelligente Systeme' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Logische Programmierung' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Multi-Agent Systems' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mobile und Cloud Computing Advanced' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Mathematische Grundlagen neuronaler Netze' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'NOSQL Datenbanken' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Spezielle Kapitel Autonome Systeme' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Software Engineering' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Weiterführende Konzepte zum Betrieb komplexer verteilter Systeme' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Übersetzerbau' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Advanced Interface Design' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Digital Forensics and Incident Response' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Datenschutz und Ethik' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Designmanagement' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Emerging Challenges in Cybersecurity Research' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Gamification' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Interaktive Kollaborative Arbeitsumgebungen' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Internet-Sicherheit A' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Internet-Sicherheit B' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Interaktive Systeme' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Interaction Design' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Malware-Analyse und Cyber Threat Intelligence' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Natural User Interfaces' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Privacy Enhancing Technologies' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Programmiermethodik und Sicherheit' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Software Reverse Engineering' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Vertiefung Digitales Marketing' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Virtuelle Welten' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Vertiefung Supply Chain Management' FROM programs WHERE name LIKE '%(M. Sc.)';
INSERT OR IGNORE INTO modules (programid, name) SELECT id, 'Zukunftstrends in der Medieninformatik' FROM programs WHERE name LIKE '%(M. Sc.)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_exams_module;
ALTER TABLE exams DROP COLUMN comment;
ALTER TABLE exams DROP COLUMN moduleid;
DROP TABLE modules;
-- +goose StatementEnd
