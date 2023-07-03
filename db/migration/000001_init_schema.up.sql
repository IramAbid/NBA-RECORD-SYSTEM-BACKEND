CREATE TABLE "students" (
                            "id" bigserial PRIMARY KEY NOT NULL,
                            "first_name" varchar(60) NOT NULL,
                            "last_name" varchar(60) NOT NULL,
                            "enrollment_no" varchar(6) NOT NULL,
                            "course" varchar (60) NOT NULL,
                            "studying_year" varchar (60) NOT NULL,
                            "email" varchar(60) NOT NULL,
                            "password" varchar(60) NOT NULL,
                            "department" varchar(60) ,
                            "faculty_no" varchar(10) ,
                            "registered_courses" TEXT [],
                            "remaining_feedbacks" TEXT[],
                            "created_at" timestamp NOT NULL DEFAULT (now()),
                            "updated_at" timestamp NOT NULL DEFAULT (now()),
);

CREATE TABLE "tokens" (
                           "id" bigserial PRIMARY KEY NOT NULL,
                           "token" varchar(255),
                           "token_hash" BYTEA,
                           "created_at" timestamp NOT NULL DEFAULT (now()),
                           "updated_at" timestamp NOT NULL DEFAULT (now()),
                           "expiry" timestamp without time zone,
);
CREATE TABLE "feedbacks" (
                           "id" bigserial PRIMARY KEY NOT NULL,
                           "enrollment_no" bigint ,
                           "subject_code" varchar(10),
                           "created_at" timestamp NOT NULL DEFAULT (now()),
);



CREATE INDEX ON "students" ("enrollment_no");

ALTER TABLE "feedbacks" ADD FOREIGN KEY ("student_id") REFERENCES "students" ("enrollment_no");


