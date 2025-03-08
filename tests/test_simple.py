def test_basic_function():
    """A simple test function"""
    assert 1 + 1 == 2

def test_another_function():
    """Another simple test function"""
    assert "hello" == "hello"

def test_failing_simple():
    """A simple failing test"""
    assert 1 + 1 == 3, "Basic math failure"
