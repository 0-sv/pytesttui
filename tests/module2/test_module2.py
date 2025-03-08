import pytest

@pytest.fixture
def sample_fixture():
    return {"key": "value"}

def test_with_fixture(sample_fixture):
    """Test using a fixture"""
    assert sample_fixture["key"] == "value"

class TestAdvanced:
    def test_advanced_method_one(self):
        """Advanced test method"""
        result = [1, 2, 3]
        assert len(result) == 3
        
    @pytest.mark.skip(reason="Example of skipped test")
    def test_skipped(self):
        """This test will be skipped"""
        assert False

def test_failing_comparison():
    """A test with a failing comparison"""
    expected = {"key1": "value1", "key2": "value2"}
    actual = {"key1": "value1", "key2": "wrong_value"}
    assert expected == actual, "Dictionary comparison failure"
