import pytest

class TestClass:
    def test_method_one(self):
        """Test method in a class"""
        assert True
        
    def test_method_two(self):
        """Another test method in the same class"""
        assert not False

def test_outside_class():
    """Test function outside of class"""
    assert 42 == 42
    
@pytest.mark.parametrize("input,expected", [
    (1, 1),
    (2, 4),
    (3, 9)
])
def test_parametrized(input, expected):
    """Parametrized test"""
    assert input * input == expected

def test_failing_with_exception():
    """A test that fails by raising an exception"""
    raise ValueError("This test intentionally raises an exception")

class TestFailingClass:
    def test_failing_method(self):
        """A failing test method in a class"""
        assert False, "This test is designed to fail"
