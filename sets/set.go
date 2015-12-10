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

// Common Go representation of a math set.
type Set map[interface{}]struct{}

func (s *Set) Insert(item interface{}) bool {
    found := s.Contains(item)
    (*s)[item] = struct{}{}
    return !found
}

func (s *Set) Remove(item interface{}) {
    delete(*s, item)
}

func (s *Set) Contains(item interface{}) bool {
    _, found := (*s)[item]
    return found
}

func (s *Set) Length() int {
    return len(*s)
}

func (s *Set) Equals(other Set) bool {
    if len(*s) != len(other) {
        return false
    }
    for i := range *s {
        if !other.Contains(i) {
            return false
        }
    }
    return true
}

func (s *Set) Clear() {
    *s = make(Set)
}

func (s *Set) Clone() Set {
    result := make(Set)
    for i := range *s {
        result.Insert(i)
    }
    return result
}

func (s *Set) Merge(other Set) {
    for i := range other {
        s.Insert(i)
    }
}

func (s *Set) Iterate() <-chan interface{} {
    ch := make(chan interface{})
    go func() {
        for i := range *s { ch <- i }
        close(ch)
    }()
    return ch
}

func (s *Set) ToSlice() []interface{} {
    result := make([]interface{}, 0, len(*s))
    for i := range *s {
        result = append(result, i)
    }
    return result
}

