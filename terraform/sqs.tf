resource "aws_iam_policy" "commandswaiting-write" {
	name = "commandswaiting-write"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["sqs:sendmessage"]
	        Effect   = "Allow"
	        Resource = "arn:aws:sqs:eu-west-2:605391140887:commandsWaiting"
	      },
	    ]
	})
}

resource "aws_iam_policy" "commandswaiting-read" {
	name = "commandswaiting-read"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["sqs:receivemessage", "sqs:deletemessage", "sqs:getqueueattributes"]
	        Effect   = "Allow"
	        Resource = "arn:aws:sqs:eu-west-2:605391140887:commandsWaiting"
	      },
	    ]
	})
}

resource "aws_iam_policy" "payloadwaitingdoctype-write" {
	name = "payloadwaitingdoctype-write"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["sqs:sendmessage"]
	        Effect   = "Allow"
	        Resource = "arn:aws:sqs:eu-west-2:605391140887:payloadWaitingDoctype"
	      },
	    ]
	})
}

resource "aws_iam_policy" "payloadwaitingdoctype-read" {
	name = "payloadwaitingdoctype-read"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["sqs:receivemessage", "sqs:deletemessage", "sqs:getqueueattributes"]
	        Effect   = "Allow"
	        Resource = "arn:aws:sqs:eu-west-2:605391140887:payloadWaitingDoctype"
	      },
	    ]
	})
}

resource "aws_iam_policy" "payloadwaitingsubmit-write" {
	name = "payloadwaitingsubmit-write"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["sqs:sendmessage"]
	        Effect   = "Allow"
	        Resource = "arn:aws:sqs:eu-west-2:605391140887:payloadWaitingSubmit"
	      },
	    ]
	})
}

resource "aws_iam_policy" "payloadwaitingsubmit-read" {
	name = "payloadwaitingsubmit-read"

	policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Action = ["sqs:receivemessage", "sqs:deletemessage", "sqs:getqueueattributes"]
	        Effect   = "Allow"
	        Resource = "arn:aws:sqs:eu-west-2:605391140887:payloadWaitingSubmit"
	      },
	    ]
	})
}