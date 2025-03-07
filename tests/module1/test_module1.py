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
