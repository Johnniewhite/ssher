"""
Built-in password generator using cryptographic randomness.
Developed by Inioluwa Adeyinka
"""

import secrets
import string


def generate_password(length: int = 20, symbols: bool = True,
                      numbers: bool = True) -> str:
    """Generate a cryptographically secure random password.

    Args:
        length: Password length (minimum 8).
        symbols: Include symbols.
        numbers: Include numbers.

    Returns:
        Generated password string.
    """
    length = max(length, 8)

    alphabet = string.ascii_letters
    required = [secrets.choice(string.ascii_lowercase),
                secrets.choice(string.ascii_uppercase)]

    if numbers:
        alphabet += string.digits
        required.append(secrets.choice(string.digits))

    if symbols:
        safe_symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?"
        alphabet += safe_symbols
        required.append(secrets.choice(safe_symbols))

    # Fill remaining length
    remaining = length - len(required)
    password_chars = required + [secrets.choice(alphabet) for _ in range(remaining)]

    # Shuffle to avoid predictable positions
    password_list = list(password_chars)
    # Fisher-Yates shuffle with secrets
    for i in range(len(password_list) - 1, 0, -1):
        j = secrets.randbelow(i + 1)
        password_list[i], password_list[j] = password_list[j], password_list[i]

    return ''.join(password_list)
