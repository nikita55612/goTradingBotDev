

XGB_BINARY = {
    'objective': 'binary:logistic',
    'eval_metric': ['logloss', 'auc', 'error', 'aucpr'],
    'booster': 'gbtree',

    'eta': 0.1,
    'gamma': 0.02,
    'max_depth': 6,
    'min_child_weight': 3,
    'subsample': 0.8,
    'colsample_bytree': 0.8,
    'lambda': 1.0,
    'alpha': 0.1,

    'tree_method': 'hist',
    'grow_policy': 'depthwise',
    'seed': 42
}
