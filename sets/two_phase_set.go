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

type TwoPhase struct {
    added   Set
    removed Set
}

func New2P() *TwoPhase {
  s := new(TwoPhase)
  s.added   = make(Set)
  s.removed = make(Set)
  return s
}

func (s *TwoPhase) Insert(item interface{}) bool {
    if s.removed.Contains(item) {
        return false
    }
    return s.added.Insert(item)
}

func (s *TwoPhase) Remove(item interface{}) bool {
    if !s.added.Contains(item) {
        return false
    }
    return s.removed.Insert(item)
}

func (s *TwoPhase) Length() int {
    return s.added.Length() - s.removed.Length()
}

func (s *TwoPhase) Contains(item interface{}) bool {
    return s.added.Contains(item) && !s.removed.Contains(item)
}

func (s *TwoPhase) Equals(other *TwoPhase) bool {
    return s.added.Equals(other.added) && s.removed.Equals(other.removed)
}

func (s *TwoPhase) Merge(other *TwoPhase) *TwoPhase {
    result := New2P()
    s.added.Merge(other.added)
    s.removed.Merge(other.removed)
    return result
}

func (s *TwoPhase) Clone() *TwoPhase {
    result := New2P()
    result.added = s.added.Clone()
    result.removed = s.removed.Clone()
    return result
}

func (s *TwoPhase) Iterate() <-chan interface{} {
    in := s.ToSet()
    return in.Iterate()
}

func (s *TwoPhase) ToSet() Set {
    result := s.added.Clone()
    for i := range s.removed {
        result.Remove(i)
    }
    return result
}

