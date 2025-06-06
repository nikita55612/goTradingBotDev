from typing import List, TypedDict
import numpy as np
import json
import csv


def read_csv(path: str) -> List[List[float]]:
    with open(path, 'r') as f:
        reader = csv.reader(f)
        return [[float(i) for i in row] for row in reader]


class SampleMetadata(TypedDict):
    index: int
    symbol: str
    client: str
    xShape: List[int]
    yShape: List[int]
    xPath: str
    yPath: str


class Dataset:
    def __init__(self, path: str, Xtype=np.float32, ytype=np.int32) -> None:
        with open(f'{path}/metadata.json', 'r') as f:
            metadata = json.load(f)

        self.Xtype = Xtype
        self.ytype = ytype

        # Параметры из metadata['params']
        self.name: str = metadata['params']['name']
        self.root_dir: str = metadata['params']['rootDir']
        self.interval: str = metadata['params']['interval']

        # Общая информация о датасете
        self.total_features: int = metadata['totalFeatures']
        self.total_signals: int = metadata['totalSignals']
        self.signals: List[str] = metadata['signals']
        self.features: List[str] = metadata['features']
        self.total_rows: int = metadata['TotalRows']
        self.total_samples: int = metadata['totalSamples']

        # Загрузка информации о samples
        self.samples: List[SampleMetadata] = metadata['samples']

        self._X_cache = None

    def __load_X(self):
        if self._X_cache is None:
            X_list = []
            for s in self.samples:
                X = np.array(read_csv(s['xPath']), self.Xtype)
                X_list.append(X)
            self._X_cache = np.vstack(X_list)

    def load(self, signal_index: int):
        y_list = []
        for s in self.samples:
            y = np.array(read_csv(s['yPath']), self.ytype)[:, signal_index]
            y_list.append(y)
        self.__load_X()
        y = np.concatenate(y_list)
        return self._X_cache, y
