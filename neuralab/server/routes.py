from aiohttp import web
import xgboost as xgb


async def ping(req: web.Request):
    return web.Response(text="pong")

async def predict(req: web.Request):
    res = {
        'data': {
            'predict': {},
            'error': '',
        },
        'status': 403,
    }
    if not req.body_exists:
        res['data']['error'] = 'request body is missing'
        return web.json_response(**res)

    try:
        body = await req.json()
    except ValueError:
        res['data']['error'] = 'invalid body format'
        return web.json_response(**res)

    for f in req.app['required_fields']:
        if f not in body:
            res['data']['error'] = f'missing required field: "{f}"'
            return web.json_response(**res)

    features = body["features"]
    markings = body['markings']

    matching_models = [
        model for model in req.app['models']
        if all(m in model for m in markings)
    ]

    if not matching_models:
        res['data']['error'] = f'no models found for markings: {", ".join(markings)}'
        res['status'] = 404
        return web.json_response(**res)

    try:
        for model in matching_models:
            if model.startswith('+xgb_'):
                dmatrix = xgb.DMatrix(features)
                model_predict = req.app['models'][model].predict(dmatrix).tolist()
                res['data']['predict'][model[1:]] = model_predict
    except Exception as e:
        res['data']['error'] = f'prediction failed: {str(e)}'
        res['status'] = 500
        return web.json_response(**res)

    if len(res['data']['predict']) == 0:
        res['data']['error'] = f'empty prediction'
        return web.json_response(**res)

    res['status'] = 200
    return web.json_response(**res)
