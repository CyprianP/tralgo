
## Projekt ## 

Backend für einen mandantenfähigen Online-Kurs-Service. Mehrere Provider, jeder verwaltet seine eigenen Kurse, Kapitel, Lektionen und Video-Metadaten. Die Trennung zwischen den Providern läuft über Row Level Security in PostgreSQL und wird damit in der Datenbank durchgesetzt.

## Technologien ##

Go als Vorgabe in der Aufgabenstellung. 

PostgreSQL wegen Row Level Security, damit lässt sich die Mandantentrennung  auf DB-Ebene lösen statt in jeder Query nach provider_id zu filtern. Die verschachtelten Kapitel laufen über eine Referenz-Spalte zum übergeordneten Kapitel (parent_chapter_id) und eine rekursive CTE.

Goyave v5 als Framework für Routing und Request/Response-Handling mit der Möglichkeit eine spätere Middleware zu implementieren für beispielsweise Authentifizierungen. 

Pgx mit pgxpool für den Datenbankzugriff. Dabei ist wichtig, dass pro Request eine TRansaktion geöffnet wird und darin app.current_provider gesetzt wird, sodass die DB-Policy auf dieser Verbindung greift. 

## Architektur ## 

- main.go startet den Server, DB-Pool und Routen 
- router_mapping hängt die HTTP - Routen an die Handler. courses/chapters/lessons/videos haben je einen Subrouter, providers als Admin-Tabelle mit provider_id im Pfad
- Handlers: Ein Handler pro Tabelle, diese nehmen Requests an, validieren die Daten und greifen auf die Datenbank zu
- tenant.go: WithTenant öffnet die Transaktion, setzt app.current_provider und führt die Query aus. Die provider_id kommt aus dem request-header. 
- Die Config Datei beinhaltet die DB-URL als Tenant user (keine Adminrechte als super user / table owner um die RLS testen zu können)
- Die database package beinhaltet das Datenbank-Schema, Beispieldaten und die RLS Policies
- Der Testordner testet die Mandantentrennung
- Volltext-Suche: subtitle_text ist über eine generierte tsvector-Spalte (subtitle_tsv) mit GIN-Index durchsuchbar modelliert
- RLS ist auf courses, chapters, lessons und video_metadata aktiv. Die Policy provider_isolation prüft provider_id = current_setting('app.current_provider').

## Starten des Projekts ##

- Lokale Datenbank "tralgo" anlegen 
- database/schema.sql ausführen
- database/sample_data.sql ausführen
- database/policies.sql ausführen
- Server starten mit go run .
- curl-Anfragen aus einem zweiten Terminal senden, z.B.: 
    curl -X PUT http://127.0.0.1:8080/courses/1 \
    -H "Content-Type: application/json" \
    -H "X-Provider-ID: 1" \
    -d '{"course_name":"Course A (edited)","course_description":"updated"}'

## Annahmen und Vereinfachungen ## 

Die providers-Tabelle hat bewusst keine RLS, sie ist die Admin-/Onboarding-Tabelle. Alle Tabellen, die die Provider selbst nutzen (courses, chapters, lessons, video_metadata), sind dagegen per RLS getrennt.

## Out of Scope ##

Authentifizierung der Tenants: Der Tenant wird aktuell über den Request-Header (X-Provider-ID) übergeben. In Produktion käme das aus einem Login bzw. einer Authentifizierung.

Benutzerverwaltung: Das Projekt läuft über eine einzelne nicht-Superuser-Rolle (tralgo) als Stellvertreter.

Volltextsuche: Das Datenmodell ist vorbereitet. subtitle_text wird über eine generierte tsvector-Spalte (subtitle_tsv) mit GIN-Index durchsuchbar gehalten, was eine Grundstruktur bietet und eine zukünftige Volltextsuche ermöglicht.

## Mögliche Erweiterungen ## 

Wiederkehrende Handler-Logik auslagern, um Boilerplate über die fünf Tabellen zu reduzieren.