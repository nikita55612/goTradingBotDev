from server import routes
from aiohttp import web
import xgboost as xgb
import os


MODELS_PATH = "models"


def run(host: str, port: int):
    app = web.Application(client_max_size=10*1024*1024)
    app['required_fields'] = [
        'features',
        'markings',
    ]
    app['models'] = {}

    for model in os.listdir(MODELS_PATH):
        model_name = model.replace('.json', '')
        if not model_name.startswith('+'):
            continue
        model_file = f'{MODELS_PATH}/{model}'
        app['models'][model_name] = xgb.Booster(model_file=model_file)

    app.add_routes(
        [
            web.get('/ping', routes.ping),
            web.post('/predict', routes.predict),
        ]
    )

    def on_start(_): return print(
        f'Neuralab servis running on http://{host}:{port}')
    web.run_app(app, host=host, port=port, print=on_start)
