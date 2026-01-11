-- +goose Up
-- +goose StatementBegin
ALTER TABLE modules ADD COLUMN alias TEXT;

UPDATE modules SET alias = 'ADS' WHERE name = 'Algorithmen und Datenstrukturen';
UPDATE modules SET alias = 'BAIN' WHERE name = 'Bachelorarbeit Informatik';
UPDATE modules SET alias = 'BSY' WHERE name = 'Betriebssysteme';
UPDATE modules SET alias = 'DBA' WHERE name = 'Datenbanksysteme';
UPDATE modules SET alias = 'EPR' WHERE name = 'Einführung in die Programmierung';
UPDATE modules SET alias = 'INS' WHERE name = 'Internet-Sprachen';
UPDATE modules SET alias = 'KBIN' WHERE name = 'Kolloquium zur Bachelorarbeit Informatik';
UPDATE modules SET alias = 'LDS' WHERE name = 'Logik und diskrete Strukturen';
UPDATE modules SET alias = 'MCI' WHERE name = 'Mensch-Computer-Interaktion';
UPDATE modules SET alias = 'MGR' WHERE name = 'Mathematische Grundlagen';
UPDATE modules SET alias = 'OPR' WHERE name = 'Objektorientierte Programmierung';
UPDATE modules SET alias = 'PXP' WHERE name = 'Praxisphase';
UPDATE modules SET alias = 'REN' WHERE name = 'Rechnernetze';
UPDATE modules SET alias = 'SLA' WHERE name = 'Statistik und Lineare Algebra';
UPDATE modules SET alias = 'SPIN' WHERE name = 'Softwareprojekt Informatik';
UPDATE modules SET alias = 'SWT' WHERE name = 'Softwaretechnik';
UPDATE modules SET alias = 'TENI' WHERE name = 'Technisches Englisch für Informatiker';
UPDATE modules SET alias = 'TGI' WHERE name = 'Technische Grundlagen der Informatik';
UPDATE modules SET alias = 'THI' WHERE name = 'Theoretische Informatik';
UPDATE modules SET alias = 'BKV' WHERE name = 'Betrieb komplexer verteilter Systeme';
UPDATE modules SET alias = 'BV' WHERE name = 'Einführung in die Bildverarbeitung';
UPDATE modules SET alias = 'DOW' WHERE name = 'Data on the Web';
UPDATE modules SET alias = 'DSP' WHERE name = 'Data Science in Practice';
UPDATE modules SET alias = 'ECC' WHERE name = 'Einführung in Cloud Computing';
UPDATE modules SET alias = 'EMI' WHERE name = 'Einführung in die medizinische Informatik';
UPDATE modules SET alias = 'ERO' WHERE name = 'Einführung in die Robotik';
UPDATE modules SET alias = 'INP' WHERE name = 'Internet-Protokolle';
UPDATE modules SET alias = 'ITR' WHERE name = 'IT-Recht';
UPDATE modules SET alias = 'ITS' WHERE name = 'Grundlagen der IT Sicherheit';
UPDATE modules SET alias = 'KBE' WHERE name = 'Komponentenbasierte Softwareentwicklung';
UPDATE modules SET alias = 'KI' WHERE name = 'Künstliche Intelligenz';
UPDATE modules SET alias = 'KGR' WHERE name = 'Knowledge Graphs';
UPDATE modules SET alias = 'MAD' WHERE name = 'Mobile Application Development';
UPDATE modules SET alias = 'MCC' WHERE name = 'Mobile und Cloud Computing';
UPDATE modules SET alias = 'MRO' WHERE name = 'Mobile Robotik';
UPDATE modules SET alias = 'PPR' WHERE name = 'Parallele Programmierung';
UPDATE modules SET alias = 'PRP' WHERE name = 'Prozedurale Programmierung';
UPDATE modules SET alias = 'PRAX' WHERE name = 'Practical Security Attacks and Exploitation';
UPDATE modules SET alias = 'SOD' WHERE name = 'Software Design';
UPDATE modules SET alias = 'BRW' WHERE name = 'Betriebliches Rechnungswesen';
UPDATE modules SET alias = 'DIM' WHERE name = 'Digitales Marketing';
UPDATE modules SET alias = 'EBWL' WHERE name = 'Einführung in die Betriebswirtschaftslehre';
UPDATE modules SET alias = 'GPM' WHERE name = 'Geschäftsprozessmanagement';
UPDATE modules SET alias = 'GWI' WHERE name = 'Grundlagen der Wirtschaftsinformatik';
UPDATE modules SET alias = 'ANS' WHERE name = 'Angewandte Netzwerksicherheit';
UPDATE modules SET alias = 'PUM' WHERE name = 'Produktion und Materialwirtschaft';
UPDATE modules SET alias = 'IGE' WHERE name = 'Informatik und Gesellschaft';
UPDATE modules SET alias = 'KMIN' WHERE name = 'Kolloquium zur Masterarbeit Informatik';
UPDATE modules SET alias = 'MAIN' WHERE name = 'Masterarbeit Informatik';
UPDATE modules SET alias = 'MPIN' WHERE name = 'Master-Projekt Informatik';
UPDATE modules SET alias = 'MSIN' WHERE name = 'Master-Seminar Informatik';
UPDATE modules SET alias = 'WVIN' WHERE name = 'Wissenschaftliche Vertiefung Informatik';
UPDATE modules SET alias = 'ASY' WHERE name = 'Autonome Systeme';
UPDATE modules SET alias = 'CV' WHERE name = 'Computer Vision';
UPDATE modules SET alias = 'DBT' WHERE name = 'Datenbanktheorie';
UPDATE modules SET alias = 'DSC' WHERE name = 'Data Science Principles';
UPDATE modules SET alias = 'EINT' WHERE name = 'Entwicklung intelligenter Systeme';
UPDATE modules SET alias = 'FCO' WHERE name = 'Future Computing';
UPDATE modules SET alias = 'FPR' WHERE name = 'Funktionale Programmierung';
UPDATE modules SET alias = 'INT' WHERE name = 'Intelligente Systeme';
UPDATE modules SET alias = 'LPR' WHERE name = 'Logische Programmierung';
UPDATE modules SET alias = 'MAS' WHERE name = 'Multi-Agent Systems';
UPDATE modules SET alias = 'MCA' WHERE name = 'Mobile und Cloud Computing Advanced';
UPDATE modules SET alias = 'MGN' WHERE name = 'Mathematische Grundlagen neuronaler Netze';
UPDATE modules SET alias = 'NSQ' WHERE name = 'NOSQL Datenbanken';
UPDATE modules SET alias = 'SAS' WHERE name = 'Spezielle Kapitel Autonome Systeme';
UPDATE modules SET alias = 'SWE' WHERE name = 'Software Engineering';
UPDATE modules SET alias = 'WKV' WHERE name = 'Weiterführende Konzepte zum Betrieb komplexer verteilter Systeme';
UPDATE modules SET alias = 'ÜSB' WHERE name = 'Übersetzerbau';
UPDATE modules SET alias = 'AID' WHERE name = 'Advanced Interface Design';
UPDATE modules SET alias = 'DFIR' WHERE name = 'Digital Forensics and Incident Response';
UPDATE modules SET alias = 'DSE' WHERE name = 'Datenschutz und Ethik';
UPDATE modules SET alias = 'DSM' WHERE name = 'Designmanagement';
UPDATE modules SET alias = 'ECCR' WHERE name = 'Emerging Challenges in Cybersecurity Research';
UPDATE modules SET alias = 'GAM' WHERE name = 'Gamification';
UPDATE modules SET alias = 'IKA' WHERE name = 'Interaktive Kollaborative Arbeitsumgebungen';
UPDATE modules SET alias = 'ISA' WHERE name = 'Internet-Sicherheit A';
UPDATE modules SET alias = 'ISB' WHERE name = 'Internet-Sicherheit B';
UPDATE modules SET alias = 'ISY' WHERE name = 'Interaktive Systeme';
UPDATE modules SET alias = 'IxD' WHERE name = 'Interaction Design';
UPDATE modules SET alias = 'MCTI' WHERE name = 'Malware-Analyse und Cyber Threat Intelligence';
UPDATE modules SET alias = 'NUI' WHERE name = 'Natural User Interfaces';
UPDATE modules SET alias = 'PET' WHERE name = 'Privacy Enhancing Technologies';
UPDATE modules SET alias = 'PMS' WHERE name = 'Programmiermethodik und Sicherheit';
UPDATE modules SET alias = 'SRE' WHERE name = 'Software Reverse Engineering';
UPDATE modules SET alias = 'VDM' WHERE name = 'Vertiefung Digitales Marketing';
UPDATE modules SET alias = 'VIW' WHERE name = 'Virtuelle Welten';
UPDATE modules SET alias = 'VSCM' WHERE name = 'Vertiefung Supply Chain Management';
UPDATE modules SET alias = 'ZTM' WHERE name = 'Zukunftstrends in der Medieninformatik';

-- Special cases for Projektmanagement
UPDATE modules SET alias = 'PM', name = 'Projektmanagement (Bachelor)' WHERE name = 'Projektmanagement' AND programid IN (SELECT id FROM programs WHERE name LIKE '%(B. Sc.)');
UPDATE modules SET alias = 'PM', name = 'Projektmanagement (Master)' WHERE name = 'Projektmanagement' AND programid IN (SELECT id FROM programs WHERE name LIKE '%(M. Sc.)');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE modules DROP COLUMN alias;
-- +goose StatementEnd
