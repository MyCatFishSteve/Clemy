provider "aws" {
  region = "${var.region}"
}

resource "aws_iam_role" "clemy_role" {
  name               = "${var.role_name}"
  assume_role_policy = "${file("assume_role_policy.json")}"
}

resource "aws_iam_policy" "clemy_policy" {
  name        = "${var.role_name}"
  description = "Grants specified permissions to the role ${var.role_name}"
  policy      = "${file("role_policy.json")}"
}

resource "aws_iam_policy_attachment" "attach_policy" {
  name       = "Policy Attachment"
  roles      = ["${aws_iam_role.clemy_role.name}"]
  policy_arn = "${aws_iam_policy.clemy_policy.arn}"
}

resource "aws_cloudwatch_event_rule" "clear_images" {
  name                = "ClearImages"
  description         = "Invokes a lambda function to clean up unused images on weekdays."
  schedule_expression = "cron(0 12 ? * MON-FRI *)"
}

resource "aws_cloudwatch_event_target" "clear_images_clemy" {
  rule = "${aws_cloudwatch_event_rule.clear_images.name}"
  arn  = "${aws_lambda_function.clemy_function.arn}"
}

resource "aws_lambda_function" "clemy_function" {
  filename         = "../build/clemy_unix.zip"
  source_code_hash = "${base64sha256(file("../build/clemy_unix.zip"))}"
  function_name    = "${var.function_name}"
  role             = "${aws_iam_role.clemy_role.arn}"
  runtime          = "go1.x"
  handler          = "build/clemy_unix"
  timeout          = "60"

  environment {
    variables = {
      CLEMY_DRY_RUN = "${var.dry_run}"
      CLEMY_VERBOSE = "${var.verbose}"
      CLEMY_MAX_AGE = "${var.max_age}"
    }
  }
}
