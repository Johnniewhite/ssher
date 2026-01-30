"""
Interactive UI composed from mixins.
Developed by Inioluwa Adeyinka
"""

from ssher.ui.base import BaseUI
from ssher.ui.display import DisplayMixin
from ssher.ui.prompts import PromptsMixin
from ssher.ui.transfer_ui import TransferUIMixin
from ssher.ui.multi_exec_ui import MultiExecUIMixin
from ssher.ui.status_ui import StatusUIMixin
from ssher.ui.help_ui import HelpUIMixin
from ssher.ui.main_loop import MainLoopMixin


class InteractiveUI(
    MainLoopMixin,
    HelpUIMixin,
    StatusUIMixin,
    MultiExecUIMixin,
    TransferUIMixin,
    PromptsMixin,
    DisplayMixin,
    BaseUI,
):
    """Full interactive UI composed from mixin classes."""
    pass
