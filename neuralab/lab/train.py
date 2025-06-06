from utils import train_test_split
from params import XGB_BINARY
from loader import Dataset
import xgboost as xgb
import numpy as np


TEST_SIZE = 0.02

params = XGB_BINARY
dataset = Dataset(r"D:\datasets\linear-trend-H1")

for signal_index, signal in enumerate(dataset.signals):
    model_name = f'xgb_{dataset.name}_{signal}'
    print(f'Model name: {model_name}')

    X, y = dataset.load(signal_index)

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=TEST_SIZE, random_state=42
    )

    dtrain = xgb.DMatrix(X_train, label=y_train)
    dvalid = xgb.DMatrix(X_test, label=y_test)

    params['device'] = 'cuda'
    params['verbosity'] = 1
    params['scale_pos_weight'] = np.sum(y == 0) / np.sum(y == 1)

    model = xgb.train(
        params,
        dtrain,
        num_boost_round=2000,
        evals=[(dvalid, 'valid')],
        verbose_eval=100,
    )

    model.save_model(f'models/{model_name}.json')


# X, y = loader.load_data(dataset, sign, 0, len(sets))
# X_train, X_valid, y_train, y_valid = train_test_split(
#     X, y, test_size=0.2, random_state=42,
# )

# train_data = lgb.Dataset(X_train, label=y_train)
# valid_data = lgb.Dataset(X_valid, label=y_valid, reference=train_data)

# model = lgb.train(
#     param,
#     train_data,
#     num_boost_round=4000,
#     valid_sets=[valid_data],
#     valid_names=['valid'],
#     callbacks=[
#         lgb.early_stopping(stopping_rounds=1000, verbose=True),
#         lgb.log_evaluation(period=100),
#         lgb.record_evaluation({}),
#     ]
# )
