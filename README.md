# go-term-editor

A simple editor I developed on in my free time. As of now this project is paused, since it took too much of my time and was not well scoped.

The project is inspired by:
- https://github.com/xi-editor/xi-editor: The implementation of a rope as a b-tree as well as many useful generic operations on it
- https://github.com/google/btree: A b-tree reference implementation in go

This project was my first experience with Go as language and I'am quite fond of it now. Though it lacks expressiveness in certain areas (like the type system or error handling), the primitives it gives you are simple to understand with no magic behind them.

To implement the terminal editor, I used https://github.com/gdamore/tcell.

## Structure
- brope: Contains a reimplementation of the rope as written in the xi-editor. It is based on the idea of a b-tree. 
- layout: Contains a crude implementation of a ui layout constraint solver. My first intention was to use linear programming for that, but but it currently is too feature poor to justify the add complexity of linear programming.
- buffer: Contains extension methods on the rope datatype. 
- btree: A copy of the go b-tree reference implementation. Would have been used as a template to implement copy on write for the b-tree rope.
