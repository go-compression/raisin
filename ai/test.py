import os

# This weirdness it to enter the engine directory, import everything, then back out
os.chdir("engine")
from engine import BenchmarkFile, Settings # type: ignore
os.chdir("..")

print("Test")
settings = Settings(WriteOutFiles=True, PrintStatus=True, PrintStats=True)
r = BenchmarkFile("arithmetic", "test.txt", settings)
print(r)
