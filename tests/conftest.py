import pytest

@pytest.fixture(scope="session")
def global_fixture():
    """A session-scoped fixture"""
    return "global data"
