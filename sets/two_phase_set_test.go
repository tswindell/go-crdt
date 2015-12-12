/*
 * Copyright (c) 2015 Tom Swindell (t.swindell@rubyx.co.uk)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package set

import "testing"

func TestNew2P(t *testing.T) {
    a := New2P()
    if a.Length() != 0 {
        t.Error("New Set should have 0 length!")
    }
}

func Test2PInsert(t *testing.T) {
    a := New2P()

    if !a.Insert(1) {
        t.Error("Failed to insert 1 into Set!")
    }

    a.Insert(2)

    if a.Insert(1) {
        t.Error("Insert returned success when attempting to insert twice!")
    }

    a.Insert(3)

    if a.Length() != 3 {
        t.Error("Set length should == 3!")
    }
}

func Test2PRemoveAndContains(t *testing.T) {
    a := New2P()

    a.Insert(1)
    a.Insert(2)
    a.Insert(3)

    if !a.Contains(2) {
        t.Error("Failed contains check in set!")
    }

    a.Remove(2)

    if a.Contains(2) {
        t.Error("Failed to remove from set!")
    }
}

func Test2PLength(t *testing.T) {
    a := New2P()

    for i := 1; i <= 10; i++ {
        a.Insert(i)
        if a.Length() != i {
            t.Errorf("Length check failed after insert! Expecting %d got %d", i, a.Length())
        }
    }

    for i := 10; i >= 1; i-- {
        a.Remove(i)
        if a.Length() != i-1 {
            t.Errorf("Length check failed after remove! Expecting %d got %d", i, a.Length())
        }
    }
}


func Test2PEquals(t *testing.T) {
    a := New2P()
    b := New2P()
    c := New2P()

    for i := 1; i <= 10; i++ { a.Insert(i); b.Insert(i) }
    for i := 1; i <= 5; i++ { c.Insert(i) }

    if !a.Equals(b) {
        t.Error("a equals b check failed!")
    }

    if a.Equals(c) {
        t.Error("a equals c check failed!")
    }
}

func Test2PClone(t *testing.T) {
    a := New2P()

    for i := 1; i <= 10; i++ {a.Insert(i)}

    if a.Length() != 10 {
        t.Errorf("Expecting length of 10 got %d", a.Length())
    }

    b := a.Clone()

    if !a.Equals(b) {
        t.Errorf("Expected a equals b!")
    }

    a.Insert(11)

    if a.Equals(b) {
        t.Errorf("Expected a not to equal b!")
    }
}

func Test2PMerge(t *testing.T) {
    a := New2P()
    b := New2P()
    c := New2P()

    for i := 1; i <= 10; i++ {
        a.Insert(i)
        c.Insert(i)
    }
    for i := 11; i <= 20; i++ {
        b.Insert(i)
        c.Insert(i)
    }

    a.Merge(b)

    if !a.Equals(c) {
        t.Error("Equals failed after merge!")
    }
}

