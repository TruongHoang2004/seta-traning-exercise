-- Create "folders" table
CREATE TABLE "public"."folders" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "folder_shares" table
CREATE TABLE "public"."folder_shares" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "folder_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "access_level" character varying(10) NOT NULL DEFAULT 'READ',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_folders_shared" FOREIGN KEY ("folder_id") REFERENCES "public"."folders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_folder_shares_folder_id" to table: "folder_shares"
CREATE INDEX "idx_folder_shares_folder_id" ON "public"."folder_shares" ("folder_id");
-- Create index "idx_folder_shares_user_id" to table: "folder_shares"
CREATE INDEX "idx_folder_shares_user_id" ON "public"."folder_shares" ("user_id");
-- Create "notes" table
CREATE TABLE "public"."notes" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "title" character varying(255) NOT NULL,
  "body" text NOT NULL,
  "folder_id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_folders_notes" FOREIGN KEY ("folder_id") REFERENCES "public"."folders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notes_folder_id" to table: "notes"
CREATE INDEX "idx_notes_folder_id" ON "public"."notes" ("folder_id");
-- Create "note_shares" table
CREATE TABLE "public"."note_shares" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "note_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "access_level" character varying(10) NOT NULL DEFAULT 'READ',
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notes_shared" FOREIGN KEY ("note_id") REFERENCES "public"."notes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_note_shares_note_id" to table: "note_shares"
CREATE INDEX "idx_note_shares_note_id" ON "public"."note_shares" ("note_id");
-- Create index "idx_note_shares_user_id" to table: "note_shares"
CREATE INDEX "idx_note_shares_user_id" ON "public"."note_shares" ("user_id");
-- Create "teams" table
CREATE TABLE "public"."teams" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "team_name" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "rosters" table
CREATE TABLE "public"."rosters" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NULL,
  "team_id" uuid NULL,
  "role" character varying(20) NULL DEFAULT 'MEMBER',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_teams_rosters" FOREIGN KEY ("team_id") REFERENCES "public"."teams" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_rosters_team_id" to table: "rosters"
CREATE INDEX "idx_rosters_team_id" ON "public"."rosters" ("team_id");
-- Create index "idx_rosters_user_id" to table: "rosters"
CREATE INDEX "idx_rosters_user_id" ON "public"."rosters" ("user_id");
-- Create index "idx_team_user" to table: "rosters"
CREATE UNIQUE INDEX "idx_team_user" ON "public"."rosters" ("user_id", "team_id");
