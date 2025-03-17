---------------------------------;
CREATE TABLE "SCHEMA"."files" (
	"id" varchar(255) NOT NULL,
	"name" text NOT NULL,
	"ext" text NOT NULL,
	"time" timestamptz NOT NULL,
	"size" bigint NOT NULL,
	"data" bytea NOT NULL,
	PRIMARY KEY ("id")
);
---------------------------------;
CREATE TABLE "SCHEMA"."jobs" (
	"id" bigserial NOT NULL,
	"cid" bigint NOT NULL,
	"rid" bigint NOT NULL,
	"name" varchar(250) NOT NULL,
	"status" varchar(1) NOT NULL,
	"locale" varchar(20) NOT NULL,
	"param" text NOT NULL,
	"state" text NOT NULL,
	"result" text NOT NULL,
	"error" text NOT NULL,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz NOT NULL,
	PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_jobs_name" ON "SCHEMA"."jobs" ("name");
---------------------------------;
CREATE TABLE "SCHEMA"."job_logs" (
	"id" bigserial NOT NULL,
	"jid" bigint NOT NULL,
	"time" timestamptz NOT NULL,
	"level" varchar(1) NOT NULL,
	"message" text NOT NULL,
	PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_job_logs_jid" ON "SCHEMA"."job_logs" ("jid");
---------------------------------;
CREATE TABLE "SCHEMA"."job_chains" (
	"id" bigserial NOT NULL,
	"name" varchar(250) NOT NULL,
	"status" varchar(1) NOT NULL,
	"states" text NOT NULL,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz NOT NULL,
	PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_job_chains_name" ON "SCHEMA"."job_chains" ("name");
---------------------------------;
CREATE TABLE "SCHEMA"."configs" (
	"name" varchar(64) NOT NULL,
	"value" text NOT NULL,
	"style" varchar(2) NOT NULL,
	"order" bigint NOT NULL,
	"required" boolean NOT NULL,
	"secret" boolean NOT NULL,
	"viewer" varchar(1) NOT NULL,
	"editor" varchar(1) NOT NULL,
	"validation" text NOT NULL,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz NOT NULL,
	PRIMARY KEY ("name")
);
---------------------------------;
CREATE TABLE "SCHEMA"."users" (
	"id" bigserial NOT NULL,
	"name" varchar(100) NOT NULL,
	"email" varchar(200) NOT NULL,
	"password" varchar(200) NOT NULL,
	"role" varchar(1) NOT NULL,
	"status" varchar(1) NOT NULL,
	"cidr" text NOT NULL,
	"secret" bigint NOT NULL,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz NOT NULL,
	PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "SCHEMA"."users" ("email");
---------------------------------;
CREATE TABLE "SCHEMA"."pets" (
	"id" bigserial NOT NULL,
	"name" varchar(100) NOT NULL,
	"gender" varchar(1) NOT NULL,
	"born_at" timestamptz NOT NULL,
	"origin" varchar(10) NOT NULL,
	"temper" varchar(1) NOT NULL,
	"habits" character(1) [],
	"amount" bigint NOT NULL,
	"price" numeric(10, 2) NOT NULL,
	"shop_name" varchar(200) NOT NULL,
	"shop_address" varchar(200) NOT NULL,
	"shop_telephone" varchar(20) NOT NULL,
	"shop_link" varchar(1000) NOT NULL,
	"description" text NOT NULL,
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz NOT NULL,
	PRIMARY KEY ("id")
);