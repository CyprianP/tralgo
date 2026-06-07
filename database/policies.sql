

-- test account as non super-user
CREATE ROLE tralgo LOGIN PASSWORD 'tralgo';
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO tralgo;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO tralgo;


-- Row level security 
ALTER TABLE courses ENABLE ROW LEVEL SECURITY;
CREATE POLICY provider_isolation ON courses
  USING (provider_id = current_setting('app.current_provider', true)::int)
  WITH CHECK (provider_id = current_setting('app.current_provider', true)::int);
ALTER TABLE courses FORCE ROW LEVEL SECURITY; -- For dev purposes to restrict table owner rights 

ALTER TABLE chapters ENABLE ROW LEVEL SECURITY;
CREATE POLICY provider_isolation ON chapters
  USING (provider_id = current_setting('app.current_provider', true)::int)
  WITH CHECK (provider_id = current_setting('app.current_provider', true)::int);
ALTER TABLE chapters FORCE ROW LEVEL SECURITY; -- For dev purposes to restrict table owner rights 

ALTER TABLE lessons ENABLE ROW LEVEL SECURITY;
CREATE POLICY provider_isolation ON lessons
  USING (provider_id = current_setting('app.current_provider', true)::int)
  WITH CHECK (provider_id = current_setting('app.current_provider', true)::int);
ALTER TABLE lessons FORCE ROW LEVEL SECURITY; -- For dev purposes to restrict table owner rights 

ALTER TABLE video_metadata ENABLE ROW LEVEL SECURITY;
CREATE POLICY provider_isolation ON video_metadata
  USING (provider_id = current_setting('app.current_provider', true)::int)
  WITH CHECK (provider_id = current_setting('app.current_provider', true)::int);
ALTER TABLE video_metadata FORCE ROW LEVEL SECURITY; -- For dev purposes to restrict table owner rights 