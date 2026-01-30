#!/usr/bin/env python3
"""
Setup script for SSHer
Developed by Inioluwa Adeyinka
"""

from setuptools import setup, find_packages

setup(
    name="ssher",
    version="3.0.0",
    author="Inioluwa Adeyinka",
    description="Ultimate SSH Configuration Manager - A powerful CLI tool for managing SSH server configurations",
    long_description=open("README.md").read() if __import__("os").path.exists("README.md") else "",
    long_description_content_type="text/markdown",
    packages=find_packages(),
    python_requires=">=3.7",
    install_requires=[
        "cryptography>=41.0.0",
        "pexpect>=4.8.0",
    ],
    entry_points={
        "console_scripts": [
            "ssher=ssher.cli:main",
        ],
    },
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "Environment :: Console",
        "Intended Audience :: Developers",
        "Intended Audience :: System Administrators",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Programming Language :: Python :: 3.13",
        "Topic :: System :: Systems Administration",
        "Topic :: Utilities",
    ],
)
