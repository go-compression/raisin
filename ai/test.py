import os
import sys

# Add the engine directory to the path of importable directories
engine_path = os.getcwd() + "/engine"
sys.path.append(engine_path)
from engine import engine

settings = engine.Settings(WriteOutFiles=False, PrintStatus=True, PrintStats=True)
r = engine.BenchmarkFile("arithmetic", "test.txt", settings)

