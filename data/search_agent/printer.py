from typing import Any

from rich.console import Console, Group
from rich.live import Live
from rich.spinner import Spinner


class Printer:
    def __init__(self, console: Console):

        self.is_jupyter = getattr(console, "is_jupyter", False)
        self.console = console

        if not self.is_jupyter:
            self.live = Live(console=console)
            self.live.start()

        self.items: dict[str, tuple[str, bool]] = {}
        self.hide_done_ids: set[str] = set()

    def end(self) -> None:
        if not self.is_jupyter:
            self.live.stop()

    def hide_done_checkmark(self, item_id: str) -> None:
        self.hide_done_ids.add(item_id)

    def update_item(
        self,
        item_id: str,
        content: str,
        is_done: bool = False,
        hide_checkmark: bool = False,
    ) -> None:
        self.items[item_id] = (content, is_done)
        if hide_checkmark:
            self.hide_done_ids.add(item_id)
        self.flush()

    def mark_item_done(self, item_id: str) -> None:
        self.items[item_id] = (self.items[item_id][0], True)
        self.flush()

    def flush(self) -> None:
        renderables: list[Any] = []
        for item_id, (content, is_done) in self.items.items():
            if is_done:
                prefix = "✅ " if item_id not in self.hide_done_ids else ""
                renderables.append(prefix + content)
            else:
                renderables.append(Spinner("dots", text=content))

        if self.is_jupyter:
            # In Jupyter, print each update in a new cell
            self.console.print(Group(*renderables))
        else:
            # In terminal, update the live display
            self.live.update(Group(*renderables))
