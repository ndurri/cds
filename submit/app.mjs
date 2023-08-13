import { s3fetch, s3put } from "./s3.mjs"
import { getToken } from "./token.mjs"
import { getAPI } from "./api.mjs"

const tokenCache = {}

const getSubmitter = docType => {
	if(docType == "Movement")
		return process.env.MOV_SUBMITTER
	else
		return process.env.DEC_SUBMITTER
}

const getValidToken = async submitter => {
	if(tokenCache[submitter] && !tokenCache[submitter].expired())
		return tokenCache[submitter]
	console.log("Requesting token for submitter ", submitter)
	if(!tokenCache[submitter])
		tokenCache[submitter] = await getToken(submitter)
	tokenCache[submitter] = await tokenCache[submitter].refresh()
	await tokenCache[submitter].save()
	console.log("Success.")
	return tokenCache[submitter]
}

export const handler = async event => {
	const bucket = event.Records[0].s3.bucket.name
	const key = event.Records[0].s3.object.key
	const uuid = key.split("/")[1]
	console.log("Processing request %s.", uuid)
	const request = await s3fetch(bucket, "requests", uuid).then(resp => resp.json())
	const submitter = getSubmitter(request.DocType)
	const token = await getValidToken(submitter)
	const api = getAPI(request.DocType)
	const content = await s3fetch(bucket, "payloads", uuid).then(resp => resp.text())
	const resp = await api.call(token.access_token, content)
	console.log("API returned %d.", resp.status)
	const respBody = await resp.text()
	request.ResponseUUID = resp.headers["x-conversation-id"]
	request.ResponseStatus = resp.status
	request.ResponseBody = respBody
	await s3put(bucket, "requests", uuid, JSON.stringify(request))
	if(resp.ok)
		await s3put(bucket, "responses", request.ResponseUUID, JSON.stringify(request))
	else
		await s3put(bucket, "failed", uuid, JSON.stringify(request))
	console.log("Finished.")	
}